package explorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/tdex-network/tdex-daemon/pkg/httputil"
)

func GetTransactionHex(hash string) (string, error) {
	url := fmt.Sprintf(
		"%s/tx/%s/hex",
		config.GetString(config.ExplorerEndpointKey),
		hash,
	)
	status, resp, err := httputil.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf(resp)
	}

	return resp, nil
}

func (e *explorer) IsTransactionConfirmed(
	txID string,
) (bool, error) {
	trxStatus, err := e.GetTransactionStatus(txID)
	if err != nil {
		return false, err
	}

	var isConfirmed bool
	switch confirmed := trxStatus["confirmed"].(type) {
	case bool:
		isConfirmed = confirmed
	}

	return isConfirmed, nil
}

func (e *explorer) GetTransactionStatus(
	txID string,
) (map[string]interface{}, error) {
	url := fmt.Sprintf(
		"%s/tx/%s/status",
		config.GetString(config.ExplorerEndpointKey),
		txID,
	)
	status, resp, err := httputil.NewHTTPRequest("GET", url, "", nil)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, err
	}

	var trxStatus map[string]interface{}
	err = json.Unmarshal([]byte(resp), &trxStatus)
	if err != nil {
		return nil, err
	}

	return trxStatus, nil
}

func BroadcastTransaction(txHex string) (string, error) {
	url := fmt.Sprintf("%s/tx", config.GetString(config.ExplorerEndpointKey))
	headers := map[string]string{
		"Content-Type": "text/plain",
	}

	status, resp, err := httputil.NewHTTPRequest(
		"POST",
		url,
		txHex,
		headers,
	)
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("electrs: %s", resp)
	}

	return resp, nil
}

func Faucet(address string) (string, error) {
	url := fmt.Sprintf("%s/faucet", config.GetString(config.ExplorerEndpointKey))
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

func Mint(address string, amount int) (string, string, error) {
	url := fmt.Sprintf("%s/mint", config.GetString(config.ExplorerEndpointKey))
	payload := map[string]interface{}{"address": address, "quantity": amount}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	respBody := map[string]interface{}{}
	err = json.Unmarshal(data, &respBody)
	if err != nil {
		return "", "", err
	}

	return respBody["txId"].(string), respBody["asset"].(string), nil
}
