package utils

import "math/rand"

// RandomString returns a random string of length n consisting of
// lowercase ASCII letters and digits. It is not suitable for security-
// or collision-sensitive uses.
func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
