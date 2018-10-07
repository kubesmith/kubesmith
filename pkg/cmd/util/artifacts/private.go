package artifacts

import (
	"path/filepath"
	"strings"
)

func convertCSVStringToArray(csv string) []string {
	found := []string{}
	tmp := strings.Split(csv, ",")

	for _, chunk := range tmp {
		chunk = strings.TrimSpace(chunk)
		found = append(found, chunk)
	}

	return found
}

func searchGlobForFiles(pattern string, files []string) []string {
	matches, err := filepath.Glob(pattern)

	if err != nil {
		return files
	}

	for _, filePath := range matches {
		files = append(files, filePath)
	}

	return files
}
