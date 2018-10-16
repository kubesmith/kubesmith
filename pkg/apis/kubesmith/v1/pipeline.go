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

// helpful functions

func (p *Pipeline) GetStages() []string {
	return p.Spec.Stages
}

func (p *Pipeline) GetTemplates() []PipelineSpecJobTemplate {
	return p.Spec.Templates
}

func (p *Pipeline) GetJobs() []PipelineSpecJob {
	return p.Spec.Jobs
}

func (p *Pipeline) GetEnvironment() []string {
	return p.Spec.Environment
}

func (p *Pipeline) GetTemplateByName(name string) (*PipelineSpecJobTemplate, error) {
	name = strings.ToLower(name)

	for _, template := range p.GetTemplates() {
		if strings.ToLower(template.Name) == name {
			return &template, nil
		}
	}

	return nil, errors.New("template does not exist")
}

func (p *Pipeline) expandJob(oldJob PipelineSpecJob) PipelineSpecJob {
	job := *oldJob.DeepCopy()
	envVars := []string{}
	artifacts := []PipelineSpecJobArtifact{}

	// if this job doesn't extend anything, we're done
	if len(job.Extends) == 0 {
		return job
	}

	// loop through the pipeline's global environment variables and add them first
	for _, env := range p.GetEnvironment() {
		envVars = append(envVars, env)
	}

	// loop through the job's specified extensions and use each extension to mutate
	// the job in the order they were specified
	for _, templateName := range job.Extends {
		template, _ := p.GetTemplateByName(templateName)

		// if there was no template by that name, keep moving on
		if template == nil {
			continue
		}

		// if the template has an image specified, overwrite the job's image
		if template.Image != "" {
			job.Image = template.Image
		}

		// add the environment variables from this template
		for _, env := range template.Environment {
			envVars = append(envVars, env)
		}

		// if the command is specified, overwrite the job's command
		if len(oldJob.Command) == 0 && len(template.Command) > 0 {
			job.Command = []string{}

			for _, value := range template.Command {
				job.Command = append(job.Command, value)
			}
		}

		// if the args are specified, overwrite the job's args
		if len(oldJob.Args) == 0 && len(template.Args) > 0 {
			job.Args = []string{}

			for _, value := range template.Args {
				job.Args = append(job.Args, value)
			}
		}

		// if the configmap data was specified, overwrite it
		if len(oldJob.ConfigMapData) == 0 && len(template.ConfigMapData) > 0 {
			job.ConfigMapData = map[string]string{}

			for key, value := range template.ConfigMapData {
				job.ConfigMapData[key] = value
			}
		}

		// add the artifacts from this template (if any were specified)
		for _, artifact := range template.Artifacts {
			artifacts = append(artifacts, artifact)
		}

		// if the template specifies an "OnlyOn" value, overwrite the current one
		// anyone using "OnlyOn" in a pipeline job needs to understand this isn't an
		// append but an overwrite
		if len(template.OnlyOn) > 0 {
			job.OnlyOn = template.OnlyOn
		}
	}

	// now that we've looped through the templates, we have all of the environment
	// variables and artifacts.

	// if the job had any environment variables specified, add them last (so they
	// overwrite any previously defined variables)
	for _, env := range job.Environment {
		envVars = append(envVars, env)
	}

	// if the job had any artifacts specified, add them last
	for _, artifact := range job.Artifacts {
		artifacts = append(artifacts, artifact)
	}

	// lastly, set the built environment variables + artifacts
	job.Environment = envVars
	job.Artifacts = artifacts

	// return the expanded job
	return job
}

func (p *Pipeline) GetExpandedJobs() []PipelineSpecJob {
	expanded := []PipelineSpecJob{}

	for _, oldJob := range p.GetJobs() {
		expanded = append(expanded, p.expandJob(oldJob))
	}

	return expanded
}

func (p *Pipeline) GetExpandedJobsForCurrentStage() []PipelineSpecJob {
	expanded := []PipelineSpecJob{}
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

func (p *Pipeline) GetCurrentPhase() PipelinePhase {
	return p.Status.Phase
}

func (p *Pipeline) SetPipelineToQueued() {
	p.Status.StageIndex = 0
	p.Status.Phase = PipelinePhaseQueued
}

func (p *Pipeline) SetPipelineToCompleted() {
	p.Status.StageIndex = len(p.GetStages())
	p.Status.Phase = PipelinePhaseCompleted
	p.Status.EndTime.Time = time.Now()
}

func (p *Pipeline) SetPipelineToRunning() {
	p.Status.StageIndex = 1
	p.Status.Phase = PipelinePhaseRunning
	p.Status.StartTime.Time = time.Now()
}

func (p *Pipeline) SetPipelineToFailed(reason string) {
	p.Status.StageIndex = len(p.GetStages())
	p.Status.Phase = PipelinePhaseFailed
	p.Status.EndTime.Time = time.Now()
	p.Status.FailureReason = reason
}

func (p *Pipeline) GetPatchFromOriginal(original Pipeline) (types.PatchType, []byte, error) {
	p.Status.LastUpdated.Time = time.Now()

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

func (p *Pipeline) GetStageIndex() int {
	return p.Status.StageIndex
}

func (p *Pipeline) GetHashID() string {
	prefix := fmt.Sprintf("%s:%s", p.GetName(), p.GetNamespace())
	hash := fnv.New32a()

	hash.Write([]byte(prefix))

	return fmt.Sprintf("%d", hash.Sum32())
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

func (p *Pipeline) IsRunning() bool {
	return p.GetCurrentPhase() == PipelinePhaseRunning
}

func (p *Pipeline) IsQueued() bool {
	return p.GetCurrentPhase() == PipelinePhaseQueued
}

func (p *Pipeline) HasCompleted() bool {
	return p.GetCurrentPhase() == PipelinePhaseCompleted
}

func (p *Pipeline) HasNoPhase() bool {
	return p.GetCurrentPhase() == PipelinePhaseEmpty
}

func (p *Pipeline) HasFailed() bool {
	return p.GetCurrentPhase() == PipelinePhaseFailed
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
