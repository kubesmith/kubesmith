package v1

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"regexp"
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

func (p *Pipeline) expandJob(oldJob PipelineJobSpec) PipelineJobSpec {
	job := *oldJob.DeepCopy()
	env := map[string]string{}
	artifacts := []PipelineJobArtifact{}

	for key, value := range p.GetEnvironment() {
		env[key] = value
	}

	for _, templateName := range job.Extends {
		template, _ := p.GetTemplateByName(templateName)
		if template == nil {
			continue
		}

		if template.Image != "" {
			job.Image = template.Image
		}

		for key, value := range template.Environment {
			env[key] = value
		}

		if len(oldJob.Command) == 0 && len(template.Command) > 0 {
			job.Command = []string{}

			for _, value := range template.Command {
				job.Command = append(job.Command, value)
			}
		}

		if len(oldJob.Args) == 0 && len(template.Args) > 0 {
			job.Args = []string{}

			for _, value := range template.Args {
				job.Args = append(job.Args, value)
			}
		}

		if len(oldJob.ConfigMapData) == 0 && len(template.ConfigMapData) > 0 {
			for key, value := range template.ConfigMapData {
				job.ConfigMapData[key] = value
			}
		}

		for _, artifact := range template.Artifacts {
			artifacts = append(artifacts, artifact)
		}

		if len(template.OnlyOn) > 0 {
			job.OnlyOn = template.OnlyOn
		}
	}

	for key, value := range oldJob.Environment {
		env[key] = value
	}

	for _, artifact := range oldJob.Artifacts {
		artifacts = append(artifacts, artifact)
	}

	job.Environment = env
	job.Artifacts = artifacts

	if len(oldJob.OnlyOn) > 0 {
		job.OnlyOn = oldJob.OnlyOn
	}

	return job
}

func (p *Pipeline) GetExpandedJobs() []PipelineJobSpec {
	expandedJobs := []PipelineJobSpec{}

	for _, oldJob := range p.GetJobs() {
		expandedJobs = append(expandedJobs, p.expandJob(oldJob))
	}

	return expandedJobs
}

func (p *Pipeline) GetExpandedJobsForCurrentStage() []PipelineJobSpec {
	expanded := []PipelineJobSpec{}
	stageName := strings.ToLower(p.GetCurrentStageName())

	if stageName == "" {
		return expanded
	}

	for _, oldJob := range p.GetJobs() {
		if strings.ToLower(oldJob.Stage) == stageName {
			expanded = append(expanded, p.expandJob(oldJob))
		}
	}

	return expanded
}

func (p *Pipeline) AdvanceCurrentStage() {
	stageIndex := p.GetStageIndex() + 1

	if stageIndex > len(p.GetStages()) {
		p.SetPhaseToSucceeded()
		return
	}

	p.Status.StageIndex = stageIndex
}

func (p *Pipeline) ValidateWorkspace() error {
	validGitURL := regexp.MustCompile(`(?:git|ssh|https?|git@[-\w.]+):(\/\/)?(.*?)(\.git)(\/?|\#[-\d\w._]+?)$`)
	if !validGitURL.MatchString(p.GetWorkspaceRepoURL()) {
		return errors.New("workspace repo url must be a valid git url")
	}

	if p.GetWorkspaceSSHSecretName() == "" {
		return errors.New("workspace ssh secret name must be specified")
	}

	if p.GetWorkspaceSSHSecretKey() == "" {
		return errors.New("workspace ssh secret key must be specified")
	}

	return nil
}

func (p *Pipeline) ValidateStages() error {
	stages := p.GetStages()
	jobs := p.GetJobs()

	if len(stages) == 0 {
		return errors.New("pipeline must have at least 1 stage specified")
	}

	for _, stage := range stages {
		stage = strings.ToLower(stage)
		hasJobs := false

		for _, job := range jobs {
			if stage == strings.ToLower(job.Stage) {
				hasJobs = true
				break
			}
		}

		if !hasJobs {
			return errors.New("each stage must have at least 1 job specified")
		}
	}

	return nil
}

func (p *Pipeline) ValidateJobs() error {
	// first, validate some basic properties of the jobs
	for _, job := range p.GetJobs() {
		// check that the stage exists
		jobStage := strings.ToLower(job.Stage)
		foundStage := false
		for _, stage := range p.GetStages() {
			if jobStage == strings.ToLower(stage) {
				foundStage = true
				break
			}
		}

		if !foundStage {
			return errors.New("job stage must be specified as a valid stage")
		}

		// check that all extends exist
		for _, extend := range job.Extends {
			foundExtend := false
			extend = strings.ToLower(extend)

			for _, extension := range p.GetTemplates() {
				if extend == strings.ToLower(extension.Name) {
					foundExtend = true
					break
				}
			}

			if !foundExtend {
				return errors.New("invalid job extension specified")
			}
		}
	}

	// now, get the expanded jobs and validate each of them
	for _, job := range p.GetExpandedJobs() {
		if err := job.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Pipeline) Validate() error {
	if err := p.ValidateWorkspace(); err != nil {
		return err
	}

	if err := p.ValidateStages(); err != nil {
		return err
	}

	return p.ValidateJobs()
}
