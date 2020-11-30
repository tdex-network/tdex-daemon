package elements

// Unspent represents an elements unspent
type Unspent struct {
	Address       string  `json:"address,omitempty"`
	Label         string  `json:"label,omitempty"`
	ScriptPubKey  string  `json:"scriptPubKey,omitempty"`
	Confirmations int64   `json:"confirmations"`
	TxID          string  `json:"txid"`
	Vout          uint32  `json:"vout"`
	Amount        float64 `json:"amount"`
	//the transaction output asset in hex
	Asset string `json:"asset,omitempty"`
	//the transaction output commitment in hex
	AmountCommitment string `json:"amountcommitment,omitempty"`
	//the transaction output asset commitment in hex
	AssetCommitment string `json:"assetcommitment,omitempty"`
	//the transaction output amount blinding factor in hex
	AmountBlinder string `json:"amountblinder,omitempty"`
	//the transaction output asset blinding factor in hex
	AssetBlinder string `json:"assetblinder,omitempty"`
}
