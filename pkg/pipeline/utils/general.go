package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"time"
)

func GetMD5(value string) string {
	hasher := md5.New()
	hasher.Write([]byte(value))

	return hex.EncodeToString(hasher.Sum(nil))
}

func GetPipelineMD5(name, namespace string) string {
	value := fmt.Sprintf("%s:%s", name, namespace)

	return GetMD5(value)
}

func GenerateRandomString(s int, letters ...string) string {
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

func Int32Ptr(i int32) *int32 {
	return &i
}
