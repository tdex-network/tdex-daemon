package swap

import (
	"testing"
)

const initialPsbtOfBob = "cHNldP8BAP1lAQIAAAAAAu1FUTUGQ6cvcZgqyRmduZP/jCOJf6CiVj9NIvClxau5AQAAAAD/////5gvIxOksvm3xwVCvGRe1AQLH8z0utX4L0e30r+VPt5cAAAAAAP////8EASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAATEtAABYAFMUoyu4AUgBlfAS1xfajCz0+UeV5AdMRFPznA5TB+dNUdQG0udNvQgI27GQZkVRWZDSIWs8tAQAADZ15iEQAABYAFMUoyu4AUgBlfAS1xfajCz0+UeV5AdMRFPznA5TB+dNUdQG0udNvQgI27GQZkVRWZDSIWs8tAQAAAAb8I6wAABYAFJJ8X9Wg477+b6SuqlazoT+3x+BHASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAFqZXAABYAFJJ8X9Wg477+b6SuqlazoT+3x+BHAAAAAAABAUIB0xEU/OcDlMH501R1AbS5029CAjbsZBmRVFZkNIhazy0BAAANpHWr8AAAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkAAQFCASWyUQcOKcoZBDzzPM1zJOLdqwPsxK4LXnfE/A5c9slaAQAAAAAF9eEAABYAFJJ8X9Wg477+b6SuqlazoT+3x+BHIgICgEVehMv0LB8AvJfc4SLP1VX1F0p6ebHBYKtzq3xbs8lHMEQCIHpIzr6p7OIGhW2PzOi6m/HKG5Gotnmt5TpylMuOSrE4AiAhxdQqlCGk4s7QaJnA2dVQc4lfWBOV3FBHaw25sM8xEQEAAAAAAA=="

func TestCore_Accept(t *testing.T) {
	t.Run("Bob can import a SwapRequest and create a SwapAccept message", func(t *testing.T) {
		messageRequest, err := Request(RequestOpts{
			AssetToSend:     USDT,
			AmountToSend:    30000000000,
			AssetToReceive:  LBTC,
			AmountToReceive: 5000000,
			PsetBase64:      initialPsbtOfAlice,
		})
		if err != nil {
			t.Errorf("Core.Request() error = %v", err)
			return
		}

		_, got, err := Accept(AcceptOpts{
			Message:    messageRequest,
			PsetBase64: initialPsbtOfBob,
		})
		if err != nil {
			t.Errorf("Core.Accept() error = %v", err)
			return
		}
		want := make([]byte, 867)
		if len(got) != len(want) {
			t.Errorf("Core.Accept() = %v, want %v", len(got), len(want))
		}

	})

}
