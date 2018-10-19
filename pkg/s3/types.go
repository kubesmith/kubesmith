package s3

import minio "github.com/minio/minio-go"

type S3Client struct {
	client *minio.Client
}
