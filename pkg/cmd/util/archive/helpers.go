package archive

import (
	"fmt"
	"strings"

	"github.com/mholt/archiver"
)

func GetSupportedFormats() []string {
	return []string{
		".zip",
		".tar",
		".tar.gz",
		".tgz",
		".tar.bz2",
		".tbz2",
		".tar.xz",
		".txz",
		".tar.lz4",
		".tlz4",
		".tar.sz",
		".tsz",
	}
}

func IsValidArchiveExtension(filePath string) bool {
	return (archiver.MatchingFormat(filePath) != nil)
}

func GetInvalidFileFormatError() error {
	formats := GetSupportedFormats()
	return fmt.Errorf("Invalid file format; valid formats are: %s", strings.Join(formats, ", "))
}

func CreateArchive(filePath string, artifacts []string) error {
	archivist := archiver.MatchingFormat(filePath)
	if archivist == nil {
		return GetInvalidFileFormatError()
	}

	if err := archivist.Make(filePath, artifacts); err != nil {
		return err
	}

	return nil
}

func ExtractArchive(archivePath string, destinationPath string) error {
	archivist := archiver.MatchingFormat(archivePath)
	if archivist == nil {
		return GetInvalidFileFormatError()
	}

	if err := archivist.Open(archivePath, destinationPath); err != nil {
		return err
	}

	return nil
}
