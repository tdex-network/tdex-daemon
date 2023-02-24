package domain

// SwapParser defines the required methods to override the default swap
// message parser, which is grpc-proto.
type SwapParser interface {
	SerializeRequest(r SwapRequest) ([]byte, int)
	SerializeAccept(
		reqMsg []byte, tx string, unblindedIns []UnblindedInput,
	) (string, []byte, int)
	SerializeComplete(accMsg []byte, tx string) (string, []byte, int)
	SerializeFail(id string, code int) (string, []byte)

	DeserializeRequest(msg []byte) *SwapRequest
	DeserializeAccept(msg []byte) *SwapAccept
	DeserializeComplete(msg []byte) *SwapComplete
	DeserializeFail(msg []byte) *SwapFail

	ParseSwapTransaction(tx string) (*SwapTransactionDetails, int)
}

type SwapTransactionDetails struct {
	PsetBase64 string
	TxHex      string
	Txid       string
}

// Swap is the data structure that represents any of the above swaps.
type Swap struct {
	Id        string
	Message   []byte
	Timestamp int64
}

type SwapRequest struct {
	Id              string
	AssetP          string
	AssetR          string
	AmountP         uint64
	AmountR         uint64
	Transaction     string
	UnblindedInputs []UnblindedInput
}

// SwapRequest is the abstracted representation of a SwapRequest message.

func (s *SwapRequest) GetId() string {
	return s.Id
}
func (s *SwapRequest) GetAssetP() string {
	return s.AssetP
}
func (s *SwapRequest) GetAmountP() uint64 {
	return s.AmountP
}
func (s *SwapRequest) GetAssetR() string {
	return s.AssetR
}
func (s *SwapRequest) GetAmountR() uint64 {
	return s.AmountR
}
func (s *SwapRequest) GetTransaction() string {
	return s.Transaction
}
func (s *SwapRequest) GetUnblindedInputs() []UnblindedInput {
	return s.UnblindedInputs
}

type SwapAccept struct {
	Id              string
	RequestId       string
	Transaction     string
	UnblindedInputs []UnblindedInput
}

func (s *SwapAccept) GetId() string {
	return s.Id
}
func (s *SwapAccept) GetRequestId() string {
	return s.RequestId
}
func (s *SwapAccept) GetTransaction() string {
	return s.Transaction
}
func (s *SwapAccept) GetUnblindedInputs() []UnblindedInput {
	return s.UnblindedInputs
}

type SwapComplete struct {
	Id          string
	AcceptId    string
	Transaction string
}

func (s *SwapComplete) GetId() string {
	return s.Id
}
func (s *SwapComplete) GetAcceptId() string {
	return s.AcceptId
}
func (s *SwapComplete) GetTransaction() string {
	return s.Transaction
}

type SwapFail struct {
	Id             string
	MessageId      string
	FailureCode    uint32
	FailureMessage string
}

func (s *SwapFail) GetId() string {
	return s.Id
}
func (s *SwapFail) GetMessageId() string {
	return s.MessageId
}
func (s *SwapFail) GetFailureCode() uint32 {
	return s.FailureCode
}
func (s *SwapFail) GetFailureMessage() string {
	return s.FailureMessage
}

type UnblindedInput struct {
	Index         uint32
	Asset         string
	Amount        uint64
	AssetBlinder  string
	AmountBlinder string
}

func (i UnblindedInput) GetIndex() uint32 {
	return i.Index
}
func (i UnblindedInput) GetAsset() string {
	return i.Asset
}
func (i UnblindedInput) GetAmount() uint64 {
	return i.Amount
}
func (i UnblindedInput) GetAssetBlinder() string {
	return i.AssetBlinder
}
func (i UnblindedInput) GetAmountBlinder() string {
	return i.AmountBlinder
}
