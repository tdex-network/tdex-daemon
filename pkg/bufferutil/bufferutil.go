package bufferutil

import (
	"encoding/hex"

	"github.com/vulpemventures/go-elements/elementsutil"
)

func AssetHashFromBytes(buffer []byte) string {
	// We remove the first byte from the buffer array that represents if confidential or unconfidential
	return hex.EncodeToString(elementsutil.ReverseBytes(buffer[1:]))
}

func AssetHashToBytes(str string) ([]byte, error) {
	buffer, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	buffer = elementsutil.ReverseBytes(buffer)
	buffer = append([]byte{0x01}, buffer...)
	return buffer, nil
}

func ValueFromBytes(buffer []byte) uint64 {
	value, _ := elementsutil.ElementsToSatoshiValue(buffer)
	return value
}

func ValueToBytes(val uint64) ([]byte, error) {
	return elementsutil.SatoshiToElementsValue(val)
}

func TxIDFromBytes(buffer []byte) string {
	return hex.EncodeToString(elementsutil.ReverseBytes(buffer))
}

func TxIDToBytes(str string) ([]byte, error) {
	buffer, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return elementsutil.ReverseBytes(buffer), nil
}

func CommitmentFromBytes(buffer []byte) string {
	return hex.EncodeToString(buffer)
}

func CommitmentToBytes(str string) ([]byte, error) {
	return hex.DecodeString(str)
}
