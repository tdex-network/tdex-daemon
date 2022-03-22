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
				end:              start.Add(24 * time.Hour), //day
				groupByTimeFrame: 4,                         //4hours
			},
			want: 24 / 4,
		},
		{
			name: "2",
			args: args{
				start:            start,
				end:              start.Add(24 * time.Hour), //day
				groupByTimeFrame: 1,                         //1 hour
			},
			want: 24 / 1,
		},
		{
			name: "3",
			args: args{
				start:            start,
				end:              start.Add(7 * 24 * time.Hour), //week
				groupByTimeFrame: 1,                             //1hour
			},
			want: 7 * 24 / 1,
		},
		{
			name: "4",
			args: args{
				start:            start,
				end:              start.Add(7 * 24 * time.Hour), //week
				groupByTimeFrame: 4,                             //4hours
			},
			want: 7 * 24 / 4,
		},
		{
			name: "5",
			args: args{
				start:            start,
				end:              start.Add(30 * 24 * time.Hour), //month
				groupByTimeFrame: 4,                              //4hours
			},
			want: 30 * 24 / 4,
		},
		{
			name: "6",
			args: args{
				start:            start,
				end:              start.Add(30 * 24 * time.Hour), //month
				groupByTimeFrame: 24,                             //24hours
			},
			want: 30 * 24 / 24,
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
