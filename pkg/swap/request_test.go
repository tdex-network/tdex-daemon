package swap

import (
	"testing"
)

const USDT = "2dcf5a8834645654911964ec3602426fd3b9b4017554d3f9c19403e7fc1411d3"
const LBTC = "5ac9f65c0efcc4775e0baec4ec03abdde22473cd3cf33c0419ca290e0751b225"
const initialPsbtOfAlice = "cHNldP8BALgCAAAAAAHtRVE1BkOnL3GYKskZnbmT/4wjiX+golY/TSLwpcWruQEAAAAA/////wIBJbJRBw4pyhkEPPM8zXMk4t2rA+zErgted8T8Dlz2yVoBAAAAAABMS0AAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkB0xEU/OcDlMH501R1AbS5029CAjbsZBmRVFZkNIhazy0BAAANnXmIRAAAFgAUxSjK7gBSAGV8BLXF9qMLPT5R5XkAAAAAAAEBQgHTERT85wOUwfnTVHUBtLnTb0ICNuxkGZFUVmQ0iFrPLQEAAA2kdavwAAAWABTFKMruAFIAZXwEtcX2ows9PlHleQAAAA=="

func TestCore_Request(t *testing.T) {
	type fields struct {
		Verbose bool
	}
	type args struct {
		opts RequestOpts
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			"Swap request",
			fields{false},
			args{RequestOpts{
				AssetToBeSent:   USDT,
				AmountToBeSent:  30000000000,
				AssetToReceive:  LBTC,
				AmountToReceive: 5000000,
				PsetBase64:      initialPsbtOfAlice,
			}},
			make([]byte, 520),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Core{
				Verbose: tt.fields.Verbose,
			}
			got, err := c.Request(tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Core.Request() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Core.Request() = %v, want %v", len(got), len(tt.want))
			}
		})
	}
}
