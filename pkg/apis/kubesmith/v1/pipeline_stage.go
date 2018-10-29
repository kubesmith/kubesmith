package v1

import (
	"encoding/json"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineStageSpec struct {
	Workspace PipelineStageWorkspace `json:"workspace"`
	Jobs      []PipelineJobSpecJob   `json:"jobs"`
}

type PipelineStageStatus struct {
	Phase           Phase       `json:"phase"`
	StartTime       metav1.Time `json:"startTime"`
	EndTime         metav1.Time `json:"endTime"`
	LastUpdatedTime metav1.Time `json:"lastUpdatedTime"`
	FailureReason   string      `json:"failureReason"`
}

type PipelineStageWorkspace struct {
	Repo    WorkspaceRepo    `json:"repo"`
	Storage WorkspaceStorage `json:"storage"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineStage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PipelineStageSpec   `json:"spec"`
	Status PipelineStageStatus `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PipelineStageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PipelineStage `json:"items"`
}

// helpers

func (p *PipelineStage) HasNoPhase() bool {
	return p.Status.Phase == PhaseEmpty
}

func (p *PipelineStage) IsRunning() bool {
	return p.Status.Phase == PhaseRunning
}

func (p *PipelineStage) HasSucceeded() bool {
	return p.Status.Phase == PhaseSucceeded
}

func (p *PipelineStage) HasFailed() bool {
	return p.Status.Phase == PhaseFailed
}

func (p *PipelineStage) SetPhaseToQueued() {
	p.Status.Phase = PhaseQueued
}

func (p *PipelineStage) SetPhaseToRunning() {
	p.Status.Phase = PhaseRunning
	p.Status.StartTime.Time = time.Now()
}

func (p *PipelineStage) SetPhaseToSucceeded() {
	p.Status.Phase = PhaseSucceeded
	p.Status.EndTime.Time = time.Now()
}

func (p *PipelineStage) SetPhaseToFailed(reason string) {
	p.Status.Phase = PhaseFailed
	p.Status.EndTime.Time = time.Now()
	p.Status.FailureReason = reason
}

func (p *PipelineStage) GetPatchFromOriginal(original PipelineStage) (types.PatchType, []byte, error) {
	p.Status.LastUpdatedTime.Time = time.Now()

	origBytes, err := json.Marshal(original)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling original pipeline stage")
	}

	updatedBytes, err := json.Marshal(p)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling updated pipeline stage")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating json merge patch for pipeline stage")
	}

	return types.MergePatchType, patchBytes, nil
}

func (p *PipelineStage) Validate() error {
	for _, job := range p.Spec.Jobs {
		if err := job.Validate(); err != nil {
			return err
		}
	}

	return nil
}
