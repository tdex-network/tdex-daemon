package config

import "testing"

func TestGetHostname(t *testing.T) {
	type args struct {
		operatorExtraIP     []string
		operatorExtraDomain []string
		connectionUrl       string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "1",
			args: args{
				operatorExtraIP:     []string{},
				operatorExtraDomain: []string{},
				connectionUrl:       "",
			},
			want:    "localhost",
			wantErr: false,
		},
		{
			name: "2",
			args: args{
				operatorExtraIP:     []string{"test1:800"},
				operatorExtraDomain: []string{"test2:800"},
				connectionUrl:       "test:8080",
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "3",
			args: args{
				operatorExtraIP:     []string{"test1:800"},
				operatorExtraDomain: []string{"test2:800"},
				connectionUrl:       "test",
			},
			want:    "test",
			wantErr: false,
		},
		{
			name: "4",
			args: args{
				operatorExtraIP:     []string{"test1:800"},
				operatorExtraDomain: []string{"test2:800"},
				connectionUrl:       "",
			},
			want:    "test2",
			wantErr: false,
		},
		{
			name: "5",
			args: args{
				operatorExtraIP:     []string{"test1:800"},
				operatorExtraDomain: []string{},
				connectionUrl:       "",
			},
			want:    "test1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getHostname(tt.args.operatorExtraIP, tt.args.operatorExtraDomain, tt.args.connectionUrl)
			if (err != nil) != tt.wantErr {
				t.Errorf("getHostname() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getHostname() got = %v, want %v", got, tt.want)
			}
		})
	}
}
