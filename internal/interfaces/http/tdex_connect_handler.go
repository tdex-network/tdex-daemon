package httpinterface

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/tdex-network/tdex-daemon/pkg/tdexdconnect"

	log "github.com/sirupsen/logrus"

	"github.com/tdex-network/tdex-daemon/internal/core/application"
)

const (
	tdexConnectHtmlTitle       = "Tdex Daemon"
	tdexConnectTemplateFile    = "web/layout.html"
	tdexConnectTemplateCssFile = "web/bulma.min.css"
)

type TdexConnectService interface {
	TdexConnectHandler(w http.ResponseWriter, req *http.Request)
}

type tdexConnect struct {
	walletUnlockerSvc application.WalletUnlockerService
	macaroonBytes     []byte
	certBytes         []byte
	serverPort        string
	macaroonPath      string
	certPath          string
}

func NewTdexConnectService(
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
	}
}

func (t *tdexConnect) TdexConnectHandler(w http.ResponseWriter, req *http.Request) {
	styleFile, err := ioutil.ReadFile(tdexConnectTemplateCssFile)
	if err != nil {
		log.Errorln(err.Error())

		http.Error(w, http.StatusText(500), 500)
		return
	}

	macaroon := t.macaroonBytes
	if t.macaroonBytes == nil {
		macaroon = readFile(t.macaroonPath)
		t.macaroonBytes = macaroon
	}

	cert := t.certBytes
	if t.certBytes == nil {
		cert = readFile(t.certPath)
		t.certBytes = cert
	}

	connectUrl, err := tdexdconnect.EncodeToString(
		t.serverPort, cert, macaroon,
	)
	if err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}

	data := Page{
		Title: tdexConnectHtmlTitle,
		Style: template.CSS(styleFile),
		Url:   template.URL(connectUrl),
	}

	//walletStatus := t.walletUnlockerSvc.IsReady(context.Background())
	//if walletStatus.Initialized {
	//	TODO check password
	//}

	if err = template.Must(template.ParseFiles(tdexConnectTemplateFile)).Execute(w, data); err != nil {
		log.Errorln(err.Error())
		http.Error(w, http.StatusText(500), 500)
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
	Title string
	Style template.CSS
	Url   template.URL
}
