package walletservice

import (
	"fmt"

	"github.com/tdex-network/tdex-daemon/config"
	"github.com/vulpemventures/go-elements/address"
)

func parseConfidentialAddress(addr string) ([]byte, []byte, error) {
	script, err := address.ToOutputScript(addr, *config.GetNetwork())
	if err != nil {
		return nil, nil, err
	}
	blindingKey, err := extractBlindingKey(addr, script)
	if err != nil {
		return nil, nil, err
	}
	return script, blindingKey, nil
}

func extractBlindingKey(addr string, script []byte) ([]byte, error) {
	addrType, _ := address.DecodeType(addr, *config.GetNetwork())
	switch addrType {
	case address.ConfidentialP2Pkh, address.ConfidentialP2Sh:
		decoded, _ := address.FromBase58(addr)
		return decoded.Data[1:34], nil
	case address.ConfidentialP2Wpkh, address.ConfidentialP2Wsh:
		decoded, _ := address.FromBlech32(addr)
		return decoded.PublicKey, nil
	default:
		return nil, fmt.Errorf("failed to extract blinding key from address '%s'", addr)
	}
}
