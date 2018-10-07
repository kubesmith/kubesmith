package artifacts

func DetectFromCSV(pathsCSV string) []string {
	fileGlobs := convertCSVStringToArray(pathsCSV)
	detectedArtifacts := []string{}

	for _, pattern := range fileGlobs {
		detectedArtifacts = searchGlobForFiles(pattern, detectedArtifacts)
	}

	return detectedArtifacts
}
