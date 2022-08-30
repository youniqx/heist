package pki

import (
	"bytes"
	"math/big"
)

const baseHex = 16

func formatSerialNumber(number *big.Int) string {
	hex := number.Text(baseHex)

	var buffer bytes.Buffer

	for i, r := range hex {
		if i > 0 && i%2 == 0 {
			buffer.WriteRune(':')
		}
		buffer.WriteRune(r)
	}

	return buffer.String()
}
