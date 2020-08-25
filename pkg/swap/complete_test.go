package swap

import (
	"testing"
)

const finalPsbtOfAlice = "cHNldP8BAP1lAQIAAAAAAu1FUTUGQ6cvcZgqyRmduZP/jCOJf6CiVj9NIvClxau5AQAAAAD/////5gvIxOksvm3xwVCvGRe1AQLH8z0utX4L0e30r+VPt5cAAAAAAP////8EASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAATEtAABYAFMUoyu4AUgBlfAS1xfajCz0+UeV5AdMRFPznA5TB+dNUdQG0udNvQgI27GQZkVRWZDSIWs8tAQAADZ15iEQAABYAFMUoyu4AUgBlfAS1xfajCz0+UeV5AdMRFPznA5TB+dNUdQG0udNvQgI27GQZkVRWZDSIWs8tAQAAAAb8I6wAABYAFJJ8X9Wg477+b6SuqlazoT+3x+BHASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAFqZXAABYAFJJ8X9Wg477+b6SuqlazoT+3x+BHAAAAAAABAUIB0xEU/OcDlMH501R1AbS5029CAjbsZBmRVFZkNIhazy0BAAANpHWr8AAAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkiAgJp6A6eYQgEnPKMfCH5c49w+9u63C62sGGTzHIJL4ZaxEcwRAIgdN5MddCGTC9hRWvUbIOREbVwhEcARauaHT4pqavp9yACIBpEdlr8hBM6e6+S6cNqDkqkqVV0JTYqKuMt5FW/abJyAQABAUIBJbJRBw4pyhkEPPM8zXMk4t2rA+zErgted8T8Dlz2yVoBAAAAAAX14QAAFgAUknxf1aDjvv5vpK6qVrOhP7fH4EciAgKARV6Ey/QsHwC8l9zhIs/VVfUXSnp5scFgq3OrfFuzyUcwRAIgekjOvqns4gaFbY/M6Lqb8cobkai2ea3lOnKUy45KsTgCICHF1CqUIaTiztBomcDZ1VBziV9YE5XcUEdrDbmwzzERAQAAAAAA"

func TestSwap_Complete(t *testing.T) {
	t.Run("Alice can import a SwapAccept message and create a SwapComplete message", func(t *testing.T) {
		s := &Swap{
			Verbose: true,
		}
		messageRequest, _ := s.Request(RequestOpts{
			AssetToBeSent:   USDT,
			AmountToBeSent:  30000000000,
			AssetToReceive:  LBTC,
			AmountToReceive: 5000000,
			PsetBase64:      initialPsbtOfAlice,
		})
		messageAccept, _ := s.Accept(AcceptOpts{
			Message:    messageRequest,
			PsetBase64: initialPsbtOfBob,
		})
		got, err := s.Complete(CompleteOpts{
			Message:    messageAccept,
			PsetBase64: finalPsbtOfAlice,
		})
		if err != nil {
			t.Errorf("Swap.Complete() error = %v ", err)
			return
		}
		want := make([]byte, 1007)
		if len(got) != len(want) {
			t.Errorf("Swap.Complete() = %v, want %v", len(got), len(want))
		}
	})
}
