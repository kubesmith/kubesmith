package v1

import "strings"

const (
	PipelineJobArtifactEventOnSuccess = "on-success"
	PipelineJobArtifactEventOnFailure = "on-failure"
)

type PipelineJobArtifactEventType string

type PipelineJobArtifact struct {
	Name  string                       `json:"name"`
	When  PipelineJobArtifactEventType `json:"when"`
	Paths []string                     `json:"paths"`
}

func (a *PipelineJobArtifact) OnSuccess() bool {
	return strings.ToLower(string(a.When)) == PipelineJobArtifactEventOnSuccess
}

func (a *PipelineJobArtifact) OnFailure() bool {
	return strings.ToLower(string(a.When)) == PipelineJobArtifactEventOnFailure
}
