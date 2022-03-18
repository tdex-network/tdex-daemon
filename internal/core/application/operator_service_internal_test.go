package application

import (
	"testing"
	"time"
)

func TestInitGroupedVolume(t *testing.T) {
	start := time.Now()
	type args struct {
		start            time.Time
		end              time.Time
		groupByTimeFrame int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "1",
			args: args{
				start:            start,
				end:              start.Add(12 * time.Hour),
				groupByTimeFrame: 4,
			},
			want: 3,
		},
		{
			name: "2",
			args: args{
				start:            start,
				end:              start.Add(12 * time.Hour),
				groupByTimeFrame: 1,
			},
			want: 12,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initGroupedVolume(tt.args.start, tt.args.end, tt.args.groupByTimeFrame); len(got) != tt.want {
				t.Errorf("initGroupedVolume() = %v, want %v", got, tt.want)
			}

		})
	}
}
