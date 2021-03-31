package elements

import "fmt"

var (
	// ErrMissingRPCHost ...
	ErrMissingRPCHost = fmt.Errorf("missing RPC host")
	// ErrMissingRPCPort ...
	ErrMissingRPCPort = fmt.Errorf("missing RPC port")
	// ErrMissingRPCUser ...
	ErrMissingRPCUser = fmt.Errorf("missing RPC user")
	// ErrMissingRPCPassword ...
	ErrMissingRPCPassword = fmt.Errorf("missing RPC password")
	// ErrInvalidRescaTimestamp ...
	ErrInvalidRescaTimestamp = fmt.Errorf("rescan timestamp must be 'now', 0 or a unix timestamp")
	// ErrInvalidAddress ...
	ErrInvalidAddress = fmt.Errorf("invalid address")
	// ErrBlindKeyNotFound ...
	ErrBlindKeyNotFound = fmt.Errorf("blindkey not found for address")
	// ErrInvalidTxJSON ...
	ErrInvalidTxJSON = fmt.Errorf("invalid tx JSON")
)
