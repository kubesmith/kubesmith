package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineSpec struct {
	Workspace   PipelineSpecWorkspace     `json:"workspace"`
	Environment []string                  `json:"environment"`
	Templates   []PipelineSpecJobTemplate `json:"templates"`
	Stages      []string                  `json:"stages"`
	Jobs        []PipelineSpecJob         `json:"jobs"`
}

type PipelineSpecWorkspace struct {
	Path    string `json:"path"`
	RepoURL string `json:"repoURL"`
}

type PipelineSpecJobTemplate struct {
	Name          string                    `json:"name"`
	Image         string                    `json:"image"`
	Environment   []string                  `json:"environment"`
	Command       []string                  `json:"command"`
	Args          []string                  `json:"args"`
	ConfigMapData map[string]string         `json:"configMapData"`
	Artifacts     []PipelineSpecJobArtifact `json:"artifacts"`
	OnlyOn        []string                  `json:"onlyOn"`
}

type PipelineSpecJob struct {
	Name          string                    `json:"name"`
	Image         string                    `json:"image"`
	Stage         string                    `json:"stage"`
	Extends       []string                  `json:"extends"`
	Environment   []string                  `json:"environment"`
	Command       []string                  `json:"command"`
	Args          []string                  `json:"args"`
	ConfigMapData map[string]string         `json:"configMapData"`
	Runner        []string                  `json:"runner"`
	AllowFailure  bool                      `json:"allowFailure"`
	Artifacts     []PipelineSpecJobArtifact `json:"artifacts"`
	OnlyOn        []string                  `json:"onlyOn"`
}

const (
	ArtifactEventOnSuccess = "on-success"
	ArtifactEventOnFailure = "on-failure"
)

type ArtifactEventType string

type PipelineSpecJobArtifact struct {
	Name  string            `json:"name"`
	When  ArtifactEventType `json:"when"`
	Paths []string          `json:"paths"`
}

const (
	PipelinePhaseQueued    = "Queued"
	PipelinePhaseRunning   = "Running"
	PipelinePhaseCompleted = "Completed"
	PipelinePhaseFailed    = "Failed"
	PipelinePhaseEmpty     = ""
)

type PipelinePhase string

type PipelineStatus struct {
	StageIndex    int                   `json:"stageIndex"`
	Phase         PipelinePhase         `json:"phase"`
	StartTime     metav1.Time           `json:"startTime"`
	EndTime       metav1.Time           `json:"endTime"`
	FailureReason string                `json:"failureReason"`
	Stages        []PipelineStatusStage `json:"stages"`
	LastUpdated   metav1.Time           `json:"lastUpdate"`
}

type PipelineStatusStage struct {
	Index int                      `json:"index"`
	Jobs  []PipelineStatusStageJob `json:"jobs"`
}

type PipelineStatusStageJobResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PipelineStatusStageJob struct {
	Index     int                              `json:"index"`
	Resource  []PipelineStatusStageJobResource `json:"resource"`
	StartTime metav1.Time                      `json:"startTime"`
	EndTime   metav1.Time                      `json:"endTime"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PipelineSpec   `json:"spec"`
	Status PipelineStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Pipeline `json:"items"`
}
