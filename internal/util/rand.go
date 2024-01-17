package util

import (
	"math/rand"
	"strings"
	"time"
)

func RandomRFC1123Name(chars int) string {
	// Define the character set for the random text.
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

	return random(chars, charset)
}

// RandomText generates random text of the specified length.
func RandomText(chars int) string {
	// Define the character set for the random text.
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	return random(chars, charset)
}

func random(chars int, charset string) string {
	// Seed the random number generator.
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Use a strings.Builder for efficient string concatenation.
	var sb strings.Builder
	length := chars
	sb.Grow(length)

	// Generate a random string of the given length.
	for i := 0; i < length; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}
