package extract

import (
	"github.com/kubesmith/kubesmith/pkg/cmd/util/archive"
)

func (o *Options) getArchivesFromS3Path() ([]string, error) {
	archives := []string{}
	files, err := o.S3.client.GetFilesFromPath(o.S3.BucketName, o.S3.Path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if archive.IsValidArchiveExtension(file) {
			archives = append(archives, file)
		}
	}

	return archives, nil
}
