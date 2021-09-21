package main

import (
	"reflect"
	"testing"
)

func TestFragmentMarketUnspents(t *testing.T) {
	type args struct {
		pair   AssetValuePair
		config map[int]int
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
				config: fragmentationMapConfig,
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
				config: fragmentationMapConfig,
			},
			want:  []uint64{2, 2, 2, 2, 2, 11, 11, 11, 16, 16, 35},
			want1: []uint64{2, 2, 2, 2, 2, 11, 11, 11, 16, 16, 38},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := fragmentUnspents(tt.args.pair, tt.args.config)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("fragmentUnspents() got = %v, numOfTrx %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("fragmentUnspents() got1 = %v, numOfTrx %v", got1, tt.want1)
			}
		})
	}
}
