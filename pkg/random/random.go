package random

import "math/rand"

const letters = "abcdefghijklmnopqrstuvwxyz"

// String returns a random string of size n
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func String(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
