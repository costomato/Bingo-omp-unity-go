package util

import (
	crand "crypto/rand"
	mrand "math/rand"
	"strings"
)

func GenerateRoomCode(n int) string {
	c, _ := crand.Prime(crand.Reader, 32)
	mrand.Seed(c.Int64())
	chars := []rune("abcdefghijklmnopqrstuvwxyz1234567890")
	length := 5
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[mrand.Intn(len(chars))])
	}
	return b.String()
}
