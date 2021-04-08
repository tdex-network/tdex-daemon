package main

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/urfave/cli/v2"
)

func TestDepositFeeCli(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	explorerFlag := cli.StringFlag{
		Name:  "explorer",
		Usage: "explorer endpoint url",
		Value: "http://127.0.0.1:3001",
	}

	app := cli.NewApp()

	app.Version = "0.0.1"
	app.Name = "tdex operator CLI"
	app.Usage = "Command line interface for tdexd daemon operators"
	app.Flags = []cli.Flag{
		&rpcFlag,
		&networkFlag,
		&explorerFlag,
	}

	app.Commands = append(
		app.Commands,
		&depositfee,
	)

	err := app.Run([]string{"", "--network=regtest", "depositfee"})
	if err != nil {
		t.Fatal(err)
	}

}

func TestFragmentFeeUnspents(t *testing.T) {
	type args struct {
		valueToBeFragmented uint64
		minFragmentValue    uint64
		maxNumOfFragments   int
	}
	tests := []struct {
		name                  string
		args                  args
		wantNumOfFragments    int
		wantLastFragmentValue uint64
	}{
		{
			name: "1",
			args: args{
				valueToBeFragmented: 100000000,
				minFragmentValue:    5000,
				maxNumOfFragments:   150,
			},
			wantNumOfFragments:    150,
			wantLastFragmentValue: 99255000,
		},
		{
			name: "2",
			args: args{
				valueToBeFragmented: 750000,
				minFragmentValue:    5000,
				maxNumOfFragments:   150,
			},
			wantNumOfFragments:    150,
			wantLastFragmentValue: 5000,
		},
		{
			name: "3",
			args: args{
				valueToBeFragmented: 749999,
				minFragmentValue:    5000,
				maxNumOfFragments:   150,
			},
			wantNumOfFragments:    149,
			wantLastFragmentValue: 9999,
		},
		{
			name: "4",
			args: args{
				valueToBeFragmented: 750030,
				minFragmentValue:    5000,
				maxNumOfFragments:   150,
			},
			wantNumOfFragments:    150,
			wantLastFragmentValue: 5030,
		},
		{
			name: "5",
			args: args{
				valueToBeFragmented: 16000,
				minFragmentValue:    5000,
				maxNumOfFragments:   150,
			},
			wantNumOfFragments:    3,
			wantLastFragmentValue: 6000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fragmentFeeUnspents(
				tt.args.valueToBeFragmented,
				tt.args.minFragmentValue,
				tt.args.maxNumOfFragments,
			)
			assert.Equal(t, len(got), tt.wantNumOfFragments)
			assert.Equal(t, got[len(got)-1], tt.wantLastFragmentValue)
			sum := uint64(0)
			for _, v := range got {
				sum += v
			}
			assert.Equal(t, sum, tt.args.valueToBeFragmented)
		})
	}
}
