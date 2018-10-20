package v1

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineSpec struct {
	Workspace   PipelineWorkspace     `json:"workspace"`
	Environment map[string]string     `json:"environment"`
	Templates   []PipelineJobTemplate `json:"templates"`
	Stages      []string              `json:"stages"`
	Jobs        []PipelineJobSpec     `json:"jobs"`
}

type PipelineWorkspace struct {
	Path    string               `json:"path"`
	RepoURL string               `json:"repoURL"`
	SSH     PipelineWorkspaceSSH `json:"ssh"`
}

type PipelineWorkspaceSSH struct {
	Secret PipelineWorkspaceSSHSecret `json:"secret"`
}

type PipelineWorkspaceSSHSecret struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type PipelineJobTemplate struct {
	Name          string                `json:"name"`
	Image         string                `json:"image"`
	Environment   map[string]string     `json:"environment"`
	Command       []string              `json:"command"`
	Args          []string              `json:"args"`
	ConfigMapData map[string]string     `json:"configMapData"`
	Artifacts     []PipelineJobArtifact `json:"artifacts"`
	OnlyOn        []string              `json:"onlyOn"`
}

type PipelineStatus struct {
	StageIndex      int         `json:"stageIndex"`
	Phase           Phase       `json:"phase"`
	StartTime       metav1.Time `json:"startTime"`
	EndTime         metav1.Time `json:"endTime"`
	LastUpdatedTime metav1.Time `json:"lastUpdatedTime"`
	FailureReason   string      `json:"failureReason"`
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

// helpers

func (p *Pipeline) GetLastUpdatedTime() metav1.Time {
	return p.Status.LastUpdatedTime
}

func (p *Pipeline) GetFailureReason() string {
	return p.Status.FailureReason
}

func (p *Pipeline) GetStartTime() metav1.Time {
	return p.Status.StartTime
}

func (p *Pipeline) GetEndTime() metav1.Time {
	return p.Status.EndTime
}

func (p *Pipeline) GetStages() []string {
	return p.Spec.Stages
}

func (p *Pipeline) GetTemplates() []PipelineJobTemplate {
	return p.Spec.Templates
}

func (p *Pipeline) GetJobs() []PipelineJobSpec {
	return p.Spec.Jobs
}

func (p *Pipeline) GetEnvironment() map[string]string {
	return p.Spec.Environment
}

func (p *Pipeline) GetPhase() Phase {
	return p.Status.Phase
}

func (p *Pipeline) GetHashID() string {
	prefix := fmt.Sprintf("%s:%s", p.GetName(), p.GetNamespace())
	hash := fnv.New32a()

	hash.Write([]byte(prefix))

	return fmt.Sprintf("%d", hash.Sum32())
}

func (p *Pipeline) GetStageIndex() int {
	return p.Status.StageIndex
}

func (p *Pipeline) GetResourcePrefix() string {
	return fmt.Sprintf("pipeline-%s", p.GetHashID())
}

func (p *Pipeline) GetResourceLabels() map[string]string {
	return map[string]string{
		"PipelineName":      p.GetName(),
		"PipelineNamespace": p.GetNamespace(),
		"PipelineID":        p.GetHashID(),
	}
}

func (p *Pipeline) GetCurrentStageName() string {
	stageIndex := p.GetStageIndex()
	stages := p.GetStages()

	if stageIndex == 0 {
		return ""
	} else if stageIndex > len(stages) {
		return ""
	}

	return stages[stageIndex-1]
}

func (p *Pipeline) GetWorkspacePath() string {
	path := p.Spec.Workspace.Path

	if path == "" {
		return "/kubesmith/workspace"
	}

	return path
}

func (p *Pipeline) GetWorkspaceRepoURL() string {
	return p.Spec.Workspace.RepoURL
}

func (p *Pipeline) GetWorkspaceSSHSecretName() string {
	return p.Spec.Workspace.SSH.Secret.Name
}

func (p *Pipeline) GetWorkspaceSSHSecretKey() string {
	return p.Spec.Workspace.SSH.Secret.Key
}

func (p *Pipeline) HasNoPhase() bool {
	return p.GetPhase() == PhaseEmpty
}

func (p *Pipeline) IsQueued() bool {
	return p.GetPhase() == PhaseQueued
}

func (p *Pipeline) IsRunning() bool {
	return p.GetPhase() == PhaseRunning
}

func (p *Pipeline) HasSucceeded() bool {
	return p.GetPhase() == PhaseSucceeded
}

func (p *Pipeline) HasFailed() bool {
	return p.GetPhase() == PhaseFailed
}

func (p *Pipeline) GetTemplateByName(name string) (*PipelineJobTemplate, error) {
	name = strings.ToLower(name)

	for _, template := range p.GetTemplates() {
		if strings.ToLower(template.Name) == name {
			return &template, nil
		}
	}

	return nil, errors.New("template does not exist")
}

func (p *Pipeline) SetPhaseToQueued() {
	p.Status.StageIndex = 0
	p.Status.Phase = PhaseQueued
}

func (p *Pipeline) SetPhaseToRunning() {
	p.Status.StageIndex = 1
	p.Status.Phase = PhaseRunning
	p.Status.StartTime.Time = time.Now()
}

func (p *Pipeline) SetPhaseToSucceeded() {
	p.Status.StageIndex = len(p.GetStages())
	p.Status.Phase = PhaseSucceeded
	p.Status.EndTime.Time = time.Now()
}

func (p *Pipeline) SetPhaseToFailed(reason string) {
	p.Status.StageIndex = len(p.GetStages())
	p.Status.Phase = PhaseFailed
	p.Status.EndTime.Time = time.Now()
	p.Status.FailureReason = reason
}

func (p *Pipeline) GetPatchFromOriginal(original Pipeline) (types.PatchType, []byte, error) {
	p.Status.LastUpdatedTime.Time = time.Now()

	origBytes, err := json.Marshal(original)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling original pipeline")
	}

	updatedBytes, err := json.Marshal(p)
	if err != nil {
		return "", nil, errors.Wrap(err, "error marshalling updated pipeline")
	}

	patchBytes, err := jsonpatch.CreateMergePatch(origBytes, updatedBytes)
	if err != nil {
		return "", nil, errors.Wrap(err, "error creating json merge patch for pipeline")
	}

	return types.MergePatchType, patchBytes, nil
}

func (p *Pipeline) AdvanceCurrentStage() {
	stageIndex := p.GetStageIndex() + 1

	if stageIndex > len(p.GetStages()) {
		p.SetPhaseToSucceeded()
		return
	}

	p.Status.StageIndex = stageIndex
}

func (p *Pipeline) Validate() error {
	// todo: finish this validation
	return nil
}
