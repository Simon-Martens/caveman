package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	mrand "math/rand"
	"time"
)

var ErrNoRandom = errors.New("no randomness available")

// Functions to generate randomness

// This tries to wait a few seconds to get a random number from the system,
// in case it fails bc the system is busy.
func CreateSecretArray(length uint, retries int) ([]byte, error) {
	b := make([]byte, length)

	for i := 0; i < retries; i++ {
		_, err := rand.Read(b)
		if err == nil {
			return b, nil
		}
		time.Sleep(time.Duration(i) * time.Second)
	}

	return nil, ErrNoRandom
}

func GenRandomUIntNotPrime() uint64 {
	n := mrand.Uint64()
	if big.NewInt(int64(n)).ProbablyPrime(0) {
		return GenRandomUIntNotPrime()
	}
	return n
}

func CreateRandomSHA256Token() (string, error) {
	// We use 256 bits of crypto/rand to generate a random token
	// We append the timestamp to make sure our seed is unique
	// Then we hash the result with sha256
	t := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(t, time.Now().Unix())

	b, err := CreateSecretArray(256, 3)
	if err != nil {
		return "", err
	}

	all := append(b, t...)
	hash := sha256.Sum256(all)

	// Sadly, 32 bits dont align to 6 bits, so there will be some padding
	bas := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	return bas, nil
}

func CreateRandomSHA512Token() (string, error) {
	// We use 256 bits of crypto/rand to generate a random token
	// We append the timestamp to make sure our seed is unique
	// Then we hash the result with sha512
	t := make([]byte, binary.MaxVarintLen64)
	_ = binary.PutVarint(t, time.Now().Unix())

	b, err := CreateSecretArray(256, 3)
	if err != nil {
		return "", err
	}

	all := append(b, t...)
	hash := sha512.Sum512(all)

	// Sadly, 64 bits dont align to 6 bits, so there will be some padding
	bas := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	return bas, nil
}
