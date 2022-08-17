package httpinterface

import (
	"context"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tdex-network/tdex-daemon/internal/core/ports"

	"github.com/tdex-network/tdex-daemon/pkg/wallet"

	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
)

const (
	tdexConnectHtmlTitle       = "TDEX Daemon"
	tdexConnectTemplateFile    = "web/layout.html"
	tdexConnectTemplateCssFile = "web/bulma.min.css"
	tdexConnectTemplateJsFile  = "web/app.js"
)

type TdexConnectService interface {
	TdexConnectHandler(w http.ResponseWriter, req *http.Request)
}

type tdexConnect struct {
	repoManager       ports.RepoManager
	walletUnlockerSvc application.WalletUnlockerService
	macaroonBytes     []byte
	certBytes         []byte
	serverPort        string
	macaroonPath      string
	certPath          string
}

func NewTdexConnectService(
	repoManager ports.RepoManager,
	walletUnlockerSvc application.WalletUnlockerService,
	macaroonPath, certPath, serverPort string,
) TdexConnectService {
	macBytes := readFile(macaroonPath)
	certBytes := readFile(certPath)

	return &tdexConnect{
		walletUnlockerSvc: walletUnlockerSvc,
		macaroonBytes:     macBytes,
		certBytes:         certBytes,
		serverPort:        serverPort,
		macaroonPath:      macaroonPath,
		certPath:          certPath,
		repoManager:       repoManager,
	}
}

func (t *tdexConnect) TdexConnectHandler(w http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	walletStatus := t.walletUnlockerSvc.IsReady(ctx)
	cert := t.certBytes
	if t.certBytes == nil {
		cert = readFile(t.certPath)
		t.certBytes = cert
	}

	connectUrl, err := tdexdconnect.EncodeToString(
		t.serverPort, cert, nil,
	)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}

	method := req.Method
	if method == http.MethodGet {
		styleFile, err := ioutil.ReadFile(tdexConnectTemplateCssFile)
		if err != nil {
			log.Errorln(err.Error())
			http.Error(w, http.StatusText(500), 500)

			return
		}

		jsFile, err := ioutil.ReadFile(tdexConnectTemplateJsFile)
		if err != nil {
			log.Errorln(err.Error())
			http.Error(w, http.StatusText(500), 500)

			return
		}

		data := Page{
			Title:               tdexConnectHtmlTitle,
			Style:               template.CSS(styleFile),
			Js:                  template.JS(jsFile),
			Url:                 template.URL(connectUrl),
			IsWalletInitialised: walletStatus.Initialized,
		}

		if err = template.Must(template.ParseFiles(tdexConnectTemplateFile)).Execute(w, data); err != nil {
			log.Errorln(err.Error())
			http.Error(w, http.StatusText(500), 500)
		}

		return
	} else {
		if walletStatus.Initialized {
			password := req.Header.Get("password")

			vault, err := t.repoManager.VaultRepository().GetOrCreateVault(ctx, nil, "", nil)
			if err != nil {
				log.Errorln(err.Error())
				http.Error(w, http.StatusText(500), 500)

				return
			}

			if _, err := wallet.Decrypt(wallet.DecryptOpts{
				CypherText: vault.EncryptedMnemonic,
				Passphrase: password,
			}); err != nil {
				log.Debugf(err.Error())
				http.Error(w, "wrong password", 500)

				return
			}

			macaroon := t.macaroonBytes
			if t.macaroonBytes == nil {
				macaroon = readFile(t.macaroonPath)
				t.macaroonBytes = macaroon
			}

			connectUrl, err = tdexdconnect.EncodeToString(
				t.serverPort, cert, nil,
			)
			if err != nil {
				log.Errorln(err.Error())
				http.Error(w, http.StatusText(500), 500)
			}
		}

		w.Write([]byte(connectUrl))
	}
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
	Title               string
	Style               template.CSS
	Js                  template.JS
	Url                 template.URL
	IsWalletInitialised bool
}
