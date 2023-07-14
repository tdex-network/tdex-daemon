package v1domain

type AddressInfo struct {
	Account        string
	Address        string
	BlindingKey    []byte
	DerivationPath string
	Script         string
}

type Wallet struct {
	EncryptedMnemonic   []byte
	PasswordHash        []byte
	BirthdayBlockHeight uint32
	RootPath            string
	NetworkName         string
	Accounts            map[string]*Account
	AccountsByLabel     map[string]string
	NextAccountIndex    uint32
}

type Account struct {
	AccountInfo
	Index                  uint32
	BirthdayBlock          uint32
	NextExternalIndex      uint
	NextInternalIndex      uint
	DerivationPathByScript map[string]string
}

type AccountInfo struct {
	Namespace      string
	Label          string
	Xpub           string
	DerivationPath string
}
