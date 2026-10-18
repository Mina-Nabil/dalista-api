package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// OTPLength is the number of digits in a generated OTP.
const OTPLength = 6

// GenerateOTP returns a cryptographically-random numeric OTP of OTPLength
// digits, zero-padded (e.g. "004213").
func GenerateOTP() (string, error) {
	max := big.NewInt(1)
	for i := 0; i < OTPLength; i++ {
		max.Mul(max, big.NewInt(10))
	}
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", OTPLength, n), nil
}
