package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/util"
	"io/ioutil"
	"net/http"
)

func GetTransactionHex(hash string) (string, error) {
	url := fmt.Sprintf(
		"%s/tx/%s/hex",
		config.GetString(config.ExplorerEndpointKey),
		hash,
	)
	status, resp, err := util.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	return resp, nil
}

func BroadcastTransaction(txHex string) (string, error) {
	url := fmt.Sprintf("%s/tx", config.GetString(config.ExplorerEndpointKey))
	headers := map[string]string{
		"Content-Type": "text/plain",
	}

	status, resp, err := util.NewHTTPRequest("POST", url, txHex, headers)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("electrs: %s", resp)
	}

	return resp, nil
}

func Faucet(address string) (string, error) {
	url := config.GetString(config.ExplorerEndpointKey) + "/faucet"
	payload := map[string]string{"address": address}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	respBody := map[string]string{}
	err = json.Unmarshal(data, &respBody)
	if err != nil {
		return "", err
	}

	return respBody["txId"], nil
}
