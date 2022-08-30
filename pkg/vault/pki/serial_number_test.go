package pki

import (
	"fmt"
	"math/big"
	"testing"
)

func parseBigInt(input string) *big.Int {
	n := new(big.Int)
	r, ok := n.SetString(input, 10)
	if !ok {
		panic(fmt.Errorf("failed to parse big int %s", input))
	}
	return r
}

func Test_formatSerialNumber(t *testing.T) {
	type args struct {
		number *big.Int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "should encode integers properly",
			args: args{number: parseBigInt("102830064483232847570845223631797449308202024258")},
			want: "12:03:0f:3f:bc:a0:d9:fa:99:a1:c0:4a:13:d3:22:5c:f7:9f:3d:42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatSerialNumber(tt.args.number); got != tt.want {
				t.Errorf("formatSerialNumber() = %v, want %v", got, tt.want)
			}
		})
	}
}
