package explorer

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetBestPairs(t *testing.T) {
	type args struct {
		items  []uint64
		target uint64
	}
	tests := []struct {
		name string
		args args
		want []uint64
	}{
		{
			name: "1",
			args: args{
				items:  []uint64{61, 61, 61, 38, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			want: []uint64{38},
		},
		{
			name: "2",
			args: args{
				items:  []uint64{61, 61, 61, 61, 61, 61, 1, 1, 1, 3},
				target: 6,
			},
			want: []uint64{3, 1, 1, 1},
		},
		{
			name: "3",
			args: args{
				items:  []uint64{61, 61},
				target: 6,
			},
			want: []uint64{61},
		},
		{
			name: "4",
			args: args{
				items:  []uint64{2, 2},
				target: 6,
			},
			want: []uint64{},
		},
		{
			name: "5",
			args: args{
				items:  []uint64{61, 1, 1, 1, 3, 56},
				target: 6,
			},
			want: []uint64{56},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sort.Slice(tt.args.items, func(i, j int) bool {
				return tt.args.items[i] > tt.args.items[j]
			})
			if got := getBestCombination(tt.args.items, tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBestPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindIndexes(t *testing.T) {
	type args struct {
		list                 []uint64
		unblindedUtxosValues []uint64
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "1",
			args: args{
				list:                 []uint64{1000},
				unblindedUtxosValues: []uint64{1000, 1000, 1000},
			},
			want: []int{0},
		},
		{
			name: "2",
			args: args{
				list:                 []uint64{1000, 1000},
				unblindedUtxosValues: []uint64{1000, 2000, 1000},
			},
			want: []int{0, 2},
		},
		{
			name: "3",
			args: args{
				list: []uint64{2000, 2000},
				unblindedUtxosValues: []uint64{1000, 2000, 1000, 2000, 2000,
					2000},
			},
			want: []int{1, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findIndexes(tt.args.list, tt.args.unblindedUtxosValues); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findIndexes() = %v, want %v", got, tt.want)
			}
		})
	}
}
