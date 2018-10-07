package s3

import (
	"fmt"

	minio "github.com/minio/minio-go"
)

func NewS3Client(host string, port int, accessKey, secretKey string, useSSL bool) (*S3Client, error) {
	client, err := minio.New(fmt.Sprintf("%s:%d", host, port), accessKey, secretKey, useSSL)
	if err != nil {
		return nil, err
	}

	return &S3Client{
		client: client,
	}, nil
}
