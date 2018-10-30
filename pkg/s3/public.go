package s3

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	minio "github.com/minio/minio-go"
)

func (s3 *S3Client) UploadFileToBucket(filePath, bucketName, remotePath string) error {
	if err := s3.CreateBucket(bucketName); err != nil {
		return err
	}

	_, err := s3.client.FPutObject(
		bucketName,
		strings.TrimLeft(fmt.Sprintf("%s/%s", remotePath, path.Base(filePath)), "/"),
		filePath,
		minio.PutObjectOptions{
			ContentType: s3.getFileType(filePath),
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (s3 *S3Client) FileExists(bucketName, filePath string) (bool, error) {
	// make sure the bucket exists
	bucketExists, err := s3.client.BucketExists(bucketName)
	if err != nil {
		return false, err
	} else if !bucketExists {
		return false, nil
	}

	// now, verify that the file exists
	if _, err = s3.client.StatObject(bucketName, filePath, minio.StatObjectOptions{}); err != nil {
		return false, nil
	}

	return true, nil
}

func (s3 *S3Client) CreateBucket(bucketName string) error {
	err := s3.client.MakeBucket(bucketName, "")

	if err != nil {
		exists, err := s3.client.BucketExists(bucketName)
		if err == nil && exists {
			return nil
		}

		return err
	}

	return nil
}

func (s3 *S3Client) DownloadFile(bucketName, remoteFilePath, localFilePath string) error {
	// open a stream to the local file
	localFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}

	// get the remote object stream
	object, err := s3.client.GetObject(bucketName, remoteFilePath, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	// copy the remote object stream to the local file steam
	if _, err = io.Copy(localFile, object); err != nil {
		return err
	}

	return nil
}
