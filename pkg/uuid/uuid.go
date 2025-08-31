package uuid

import (
	"crypto/rand"
	"fmt"
)

func New() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Set version 4 (random).
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant (RFC4122).
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
