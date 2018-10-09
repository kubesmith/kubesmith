package templates

import (
	"crypto/rand"
	mathrand "math/rand"
	"time"
)

func generateRandomString(s int, letters ...string) string {
	randomFactor := make([]byte, 1)
	_, err := rand.Read(randomFactor)
	if err != nil {
		return ""
	}

	mathrand.Seed(time.Now().UnixNano() * int64(randomFactor[0]))
	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
	if len(letters) > 0 {
		letterRunes = []rune(letters[0])
	}

	b := make([]rune, s)
	for i := range b {
		b[i] = letterRunes[mathrand.Intn(len(letterRunes))]
	}

	return string(b)
}

func int32Ptr(i int32) *int32 {
	return &i
}
