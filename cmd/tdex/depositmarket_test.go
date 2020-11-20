package main

import (
	"github.com/urfave/cli/v2"
	"reflect"
	"testing"
)

func TestFragmentation(t *testing.T) {
	type args struct {
		pair AssetValuePair
	}
	tests := []struct {
		name  string
		args  args
		want  []uint64
		want1 []uint64
	}{
		{
			name: "test1",
			args: args{
				pair: AssetValuePair{
					BaseValue:  100000000,
					QuoteValue: 400000000000000,
				},
			},
			want: []uint64{
				2000000,
				2000000,
				2000000,
				2000000,
				2000000,
				10000000,
				10000000,
				10000000,
				15000000,
				15000000,
				30000000,
			},
			want1: []uint64{
				8000000000000,
				8000000000000,
				8000000000000,
				8000000000000,
				8000000000000,
				40000000000000,
				40000000000000,
				40000000000000,
				60000000000000,
				60000000000000,
				120000000000000,
			},
		},
		{
			name: "test2",
			args: args{
				pair: AssetValuePair{
					BaseValue:  110,
					QuoteValue: 113,
				},
			},
			want:  []uint64{2, 2, 2, 2, 2, 11, 11, 11, 16, 16, 35},
			want1: []uint64{2, 2, 2, 2, 2, 11, 11, 11, 16, 16, 38},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := fragmentUnspents(tt.args.pair)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fragmentUnspents() got = %v, numOfTrx %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("fragmentUnspents() got1 = %v, numOfTrx %v", got1, tt.want1)
			}
		})
	}
}

func TestDepositMarketCli(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	app := cli.NewApp()

	app.Version = "0.0.1"
	app.Name = "tdex operator CLI"
	app.Usage = "Command line interface for tdexd daemon operators"
	app.Flags = []cli.Flag{
		&rpcFlag,
		&networkFlag,
	}

	app.Commands = append(
		app.Commands,
		&depositmarket,
	)

	err := app.Run([]string{"", "depositmarket"})
	if err != nil {
		t.Fatal(err)
	}

}
