package cachet

import "testing"

func TestCheckTCPPortAlive(t *testing.T) {

	timeoutSecond := int64(5)

	type args struct {
		host    string
		port    string
		timeout int64
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"test 114 DNS",
			args{
				"114.114.114.114",
				"53",
				timeoutSecond,
			},
			true,
		},
		{
			"test port with hostname",
			args{
				"baidu.com",
				"443",
				timeoutSecond,
			},
			true,
		},
		{
			"test not work port with ip (it should failed)",
			args{
				"220.167.78.233",
				"600001",
				timeoutSecond,
			},
			false,
		},
		{
			"test not work port with hostname (it should failed)",
			args{
				"baidu.com",
				"444",
				timeoutSecond,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CheckTCPPortAlive(tt.args.host, tt.args.port, tt.args.timeout); got != tt.want {
				t.Errorf("CheckTCPPortAlive() = %v, want %v", got, tt.want)
			}
		})
	}
}
