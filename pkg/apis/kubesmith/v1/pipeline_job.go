package v1

import (
	"encoding/json"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineJobSpec struct {
	Name          string                `json:"name"`
	Image         string                `json:"image"`
	Stage         string                `json:"stage"`
	Extends       []string              `json:"extends"`
	Environment   map[string]string     `json:"environment"`
	Command       []string              `json:"command"`
	Args          []string              `json:"args"`
	ConfigMapData map[string]string     `json:"configMapData"`
	Runner        []string              `json:"runner"`
	AllowFailure  bool                  `json:"allowFailure"`
	Artifacts     []PipelineJobArtifact `json:"artifacts"`
	OnlyOn        []string              `json:"onlyOn"`
}

const (
	ArtifactEventOnSuccess = "on-success"
	ArtifactEventOnFailure = "on-failure"
)

type ArtifactEventType string

type PipelineJobArtifact struct {
	Name  string            `json:"name"`
	When  ArtifactEventType `json:"when"`
	Paths []string          `json:"paths"`
}

const (
	PhaseEmpty     = ""
	PhaseQueued    = "Queued"
	PhaseRunning   = "Running"
	PhaseSucceeded = "Succeeded"
	PhaseFailed    = "Failed"
)

type Phase string

type PipelineJobStatus struct {
	Phase           Phase       `json:"phase"`
	StartTime       metav1.Time `json:"startTime"`
	EndTime         metav1.Time `json:"endTime"`
	LastUpdatedTime metav1.Time `json:"lastUpdatedTime"`
	FailureReason   string      `json:"failureReason"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PipelineJobSpec   `json:"spec"`
	Status PipelineJobStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PipelineJob `json:"items"`
}

// helpers

func (p *PipelineJob) GetLastUpdatedTime() metav1.Time {
	return p.Status.LastUpdatedTime
}

func (p *PipelineJob) GetFailureReason() string {
	return p.Status.FailureReason
}

func (p *PipelineJob) GetStartTime() metav1.Time {
	return p.Status.StartTime
}

func (p *PipelineJob) GetEndTime() metav1.Time {
	return p.Status.EndTime
}

func (p *PipelineJob) GetPhase() Phase {
	return p.Status.Phase
}

func (p *PipelineJob) GetJobName() string {
	return p.Spec.Name
}

func (p *PipelineJob) GetImage() string {
	return p.Spec.Image
}

func (p *PipelineJob) GetStage() string {
	return p.Spec.Stage
}

func (p *PipelineJob) GetExtends() []string {
	return p.Spec.Extends
}

func (p *PipelineJob) GetEnvironment() map[string]string {
	return p.Spec.Environment
}

func (p *PipelineJob) GetCommand() []string {
	if len(p.Spec.Command) > 0 {
		return p.Spec.Command
	}

	return []string{"/bin/sh", "-x", "/kubesmith/scripts/pipeline-script.sh"}
}

func (p *PipelineJob) GetArgs() []string {
	return p.Spec.Args
}

func (p *PipelineJob) GetConfigMapData() map[string]string {
	if len(p.Spec.ConfigMapData) > 0 {
		return p.Spec.ConfigMapData
	}

	return map[string]string{"pipeline-script.sh": strings.Join(p.GetRunner(), "\n")}
}

func (p *PipelineJob) GetRunner() []string {
	return p.Spec.Runner
}

func (p *PipelineJob) IsAllowedToFail() bool {
	return p.Spec.AllowFailure == true
}

func (p *PipelineJob) GetArtifacts() []PipelineJobArtifact {
	return p.Spec.Artifacts
}

func (p *PipelineJob) GetOnlyOn() []string {
	return p.Spec.OnlyOn
}

func (p *PipelineJob) HasNoPhase() bool {
	return p.GetPhase() == PhaseEmpty
}

func (p *PipelineJob) IsQueued() bool {
	return p.GetPhase() == PhaseQueued
}

func (p *PipelineJob) IsRunning() bool {
	return p.GetPhase() == PhaseRunning
}

func (p *PipelineJob) HasSucceeded() bool {
	return p.GetPhase() == PhaseSucceeded
}

func (p *PipelineJob) HasFailed() bool {
	return p.GetPhase() == PhaseFailed
}

func (p *PipelineJob) SetPhaseToQueued() {
	p.Status.Phase = PhaseQueued
}

func (p *PipelineJob) SetPhaseToRunning() {
	p.Status.Phase = PhaseRunning
	p.Status.StartTime.Time = time.Now()
}

func (p *PipelineJob) SetPhaseToSucceeded() {
	p.Status.Phase = PhaseSucceeded
	p.Status.EndTime.Time = time.Now()
}

func (p *PipelineJob) SetPhaseToFailed(reason string) {
	p.Status.Phase = PhaseFailed
	p.Status.EndTime.Time = time.Now()
	p.Status.FailureReason = reason
}

func (p *PipelineJob) GetPatchFromOriginal(original PipelineJob) (types.PatchType, []byte, error) {
	p.Status.LastUpdatedTime.Time = time.Now()

	origBytes, err := json.Marshal(original)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling original pipeline job")
	}

	updatedBytes, err := json.Marshal(p)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling updated pipeline job")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating json merge patch for pipeline job")
	}

	return types.MergePatchType, patchBytes, nil
}

func (p *PipelineJobSpec) Validate() error {
	if p.Name == "" {
		return errors.New("job name must not be empty")
	}

	if p.Image == "" {
		return errors.New("job image must not be empty")
	}

	if p.Stage == "" {
		return errors.New("job stage must be specified")
	}

	hasCommands := len(p.Command) > 0 || len(p.Args) > 0
	hasRunner := len(p.Runner) > 0

	if hasCommands && hasRunner {
		return errors.New("job must have either command/args or runner specified; not both")
	}

	return nil
}

func (p *PipelineJob) Validate() error {
	return p.Spec.Validate()
}
