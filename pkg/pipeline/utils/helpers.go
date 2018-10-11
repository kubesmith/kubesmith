package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"time"

	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

func GetPipelineMD5(pipeline *api.Pipeline) string {
	hasher := md5.New()
	hasher.Write([]byte(pipeline.Name))

	return hex.EncodeToString(hasher.Sum(nil))
}

func GetPipelineResourcePrefix(pipeline *api.Pipeline) string {
	return fmt.Sprintf("pipeline-%s", GetPipelineMD5(pipeline))
}

func GetPipelineResourceLabels(pipeline *api.Pipeline) map[string]string {
	labels := map[string]string{
		"PipelineName": pipeline.Name,
		"PipelineMD5":  GetPipelineMD5(pipeline),
	}

	return labels
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
