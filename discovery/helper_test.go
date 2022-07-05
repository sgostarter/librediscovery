package discovery

import "testing"

func TestParseDiscoveryServerName(t *testing.T) {
	type args struct {
		n string
	}
	tests := []struct {
		name      string
		args      args
		wantT     string
		wantName  string
		wantIndex string
		wantErr   bool
	}{
		{
			"test",
			args{
				n: "grpc:server",
			},
			"grpc",
			"server",
			"",
			false,
		},
		{
			"test",
			args{
				n: "grpc:server:1",
			},
			"grpc",
			"server",
			"1",
			false,
		},
		{
			"test",
			args{
				n: "grpc:server:1:2",
			},
			"grpc",
			"server:1",
			"2",
			false,
		},
		{
			"test",
			args{
				n: "grpcb:server:1:2",
			},
			"grpcb",
			"",
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, gotName, gotIndex, err := ParseDiscoveryServerName(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDiscoveryServerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotT != tt.wantT {
				t.Errorf("ParseDiscoveryServerName() gotT = %v, want %v", gotT, tt.wantT)
			}
			if gotName != tt.wantName {
				t.Errorf("ParseDiscoveryServerName() gotName = %v, want %v", gotName, tt.wantName)
			}
			if gotIndex != tt.wantIndex {
				t.Errorf("ParseDiscoveryServerName() gotIndex = %v, want %v", gotIndex, tt.wantIndex)
			}
		})
	}
}
