package v1

type PipelineJobArtifactEventType string

type PipelineJobArtifacts struct {
	OnSuccess []string `json:"onSuccess"`
	OnFail    []string `json:"onFail"`
}
