package httpinterface

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/btcsuite/btcutil"
	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"

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
	macaroonPath, certPath, addr, protocol string,
) (TdexConnectService, error) {
	macBytes := readFile(macaroonPath)
	certBytes := readFile(certPath)

	if addr == "" {
		return nil, fmt.Errorf("tdexconnet: missing listening address")
	}

	p := protocol
	if len(certBytes) > 0 {
		if protocol == httpsProtocol {
			p = httpsProtocol
		} else {
			return nil, fmt.Errorf(
				"tdexdconnect: proto must be %s if cert is given, got %s",
				httpsProtocol, protocol,
			)
		}
	} else {
		if protocol == httpsProtocol {
			return nil, fmt.Errorf(
				"tdexdconnect: proto must be %s if cert is not given, got %s",
				httpProtocol, protocol,
			)
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
		serverAddress:     addr,
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		log.Debugln("http: basic auth not provided")
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
	if username != "tdex" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
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

	pwdHash := btcutil.Hash160([]byte(password))
	if !bytes.Equal(vault.PassphraseHash, pwdHash) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
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
