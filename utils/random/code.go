package random

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %v", err)
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

func Generate6DigitCode() (string, error) {
	return GenerateCode(6)
}
