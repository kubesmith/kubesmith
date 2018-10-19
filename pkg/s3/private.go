package s3

import (
	"io/ioutil"

	filetype "gopkg.in/h2non/filetype.v1"
)

func (s3 *S3Client) getFileType(filePath string) string {
	buf, _ := ioutil.ReadFile(filePath)

	kind, unknown := filetype.Match(buf)
	if unknown != nil {
		return ""
	}

	return kind.MIME.Value
}
