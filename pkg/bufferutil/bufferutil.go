package bufferutil

import (
	"encoding/hex"

	"github.com/vulpemventures/go-elements/confidential"
)

// reverseBytes returns a copy of the given byte slice with elems in reverse order.
func ReverseBytes(buf []byte) []byte {
	if len(buf) < 1 {
		return buf
	}
	tmp := make([]byte, len(buf))
	copy(tmp, buf)
	for i := len(tmp)/2 - 1; i >= 0; i-- {
		j := len(tmp) - 1 - i
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}
	return tmp
}

func AssetHashFromBytes(buffer []byte) string {
	// We remove the first byte from the buffer array that represents if confidential or unconfidential
	return hex.EncodeToString(ReverseBytes(buffer[1:]))
}

func ValueFromBytes(buffer []byte) uint64 {
	var elementsValue [9]byte
	copy(elementsValue[:], buffer[0:9])
	value, _ := confidential.ElementsToSatoshiValue(elementsValue)
	return value
}

func AssetHashToBytes(str string) ([]byte, error) {
	buffer, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	buffer = ReverseBytes(buffer)
	buffer = append([]byte{0x01}, buffer...)
	return buffer, nil
}

func ValueToBytes(val uint64) ([]byte, error) {
	buffer, err := confidential.SatoshiToElementsValue(val)
	if err != nil {
		return nil, err
	}
	return buffer[:], nil
}

func TxIDFromBytes(buffer []byte) string {
	return hex.EncodeToString(ReverseBytes(buffer))
}
