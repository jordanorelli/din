package dinutil

import (
	crand "crypto/rand"
	"encoding/base64"
	"io"
	"math/rand"
	"time"
)

func init() { rand.Seed(time.Now().Unix()) }

// generates a pseudorandom string of length n that is composed of alphanumeric
// characters.
func RandomString(n int) string {
	var alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	buf := make([]byte, n)
	for i := 0; i < len(buf); i++ {
		buf[i] = alpha[rand.Intn(len(alpha)-1)]
	}
	return string(buf)
}

// generates a pseudorandom alphanumeric string of length n that can be easily
// read and entered by a human.  The following characters are omitted to avoid
// user error: i, g, l, o, s, u, v, z, G, I, L, O, S, U, V, Z, 0, 1, 2, 5, 6, 9
func HumanRandom(n int) string {
	var alpha = "abcdefhjkmnqrtwxyABCDEFHJKMNPQRTWXY3478"
	buf := make([]byte, n)
	for i := 0; i < len(buf); i++ {
		buf[i] = alpha[rand.Intn(len(alpha)-1)]
	}
	return string(buf)
}

// generates base64 string of cryptographically random data.
func CryptoRand(n int) string {
	m := base64.StdEncoding.DecodedLen(n)
	b := make([]byte, m)
	// this might be an error.
	io.ReadFull(crand.Reader, b)
	return base64.StdEncoding.EncodeToString(b)
}

// generates a URL-safe base64 string of cryptographically random data.
func CryptoRandURL(n int) string {
	m := base64.URLEncoding.DecodedLen(n)
	b := make([]byte, m)
	// this might be an error.
	io.ReadFull(crand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}
