package httpinterface

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"
	"github.com/tdex-network/tdex-daemon/pkg/wallet"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
)

const (
	templateFile  = "web/layout.html"
	httpProtocol  = "http"
	httpsProtocol = "https"
)

type TdexConnectService interface {
	RootHandler(w http.ResponseWriter, req *http.Request)
	AuthHandler(w http.ResponseWriter, req *http.Request)
}

type tdexConnect struct {
	repoManager       ports.RepoManager
	walletUnlockerSvc application.WalletUnlockerService
	macaroonBytes     []byte
	certBytes         []byte
	macaroonPath      string
	certPath          string
	serverAddress     string
	protocol          string
}

func NewTdexConnectService(
	repoManager ports.RepoManager,
	walletUnlockerSvc application.WalletUnlockerService,
	macaroonPath, certPath, serverPort, hostname, protocol string,
) (TdexConnectService, error) {
	macBytes := readFile(macaroonPath)
	certBytes := readFile(certPath)

	p := protocol
	if len(certBytes) > 0 {
		if protocol == "" || protocol == httpsProtocol {
			p = httpsProtocol
		} else if protocol == httpProtocol {
			return nil, errors.New("tdexdconnect: proto=http is not valid if cert is given")
		}
	}

	return &tdexConnect{
		walletUnlockerSvc: walletUnlockerSvc,
		macaroonBytes:     macBytes,
		certBytes:         certBytes,
		macaroonPath:      macaroonPath,
		certPath:          certPath,
		repoManager:       repoManager,
		protocol:          p,
		serverAddress:     fmt.Sprintf("%s:%s", hostname, serverPort),
	}, nil
}

func (t *tdexConnect) RootHandler(w http.ResponseWriter, req *http.Request) {
	data := Page{
		Title: "TDEX Daemon",
	}

	if err := template.Must(template.ParseFiles(templateFile)).Execute(w, data); err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (t *tdexConnect) AuthHandler(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()

	cert := t.certBytes
	if t.certBytes == nil {
		cert = readFile(t.certPath)
		t.certBytes = cert
	}

	// start building the TDEXD connect URL
	connectUrl, err := tdexdconnect.EncodeToString(
		t.serverAddress, t.protocol, cert, nil,
	)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	walletStatus := t.walletUnlockerSvc.IsReady(ctx)
	if !walletStatus.Initialized {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(connectUrl))
		return
	}

	// if is initialized then we need to check auth before appending the macaroon
	username, password, ok := req.BasicAuth()
	if !ok {
		log.Debugln("http: basic auth not provided")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if username != "tdex" {
		log.Debugln("http: invalid username")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	vault, err := t.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if _, err := wallet.Decrypt(wallet.DecryptOpts{
		CypherText: vault.EncryptedMnemonic,
		Passphrase: password,
	}); err != nil {
		log.Debugln("http: invalid password")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	// serialize the macaroon and append it to the connect URL
	macaroon := t.macaroonBytes
	if t.macaroonBytes == nil {
		macaroon = readFile(t.macaroonPath)
		t.macaroonBytes = macaroon
	}

	// append the macaroon
	connectUrl, err = tdexdconnect.EncodeToString(
		t.serverAddress, t.protocol, cert, macaroon,
	)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write([]byte(connectUrl))
}

func readFile(filePath string) []byte {
	if filePath == "" {
		return nil
	}

	if _, err := os.Stat(filePath); err != nil {
		return nil
	}

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil
	}

	return file
}

type Page struct {
	Title string
}
