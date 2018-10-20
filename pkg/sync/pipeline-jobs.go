package sync

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

type PipelineJobSyncAction struct {
	action SyncActionType
	job    api.PipelineJob
}

func (a *PipelineJobSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *PipelineJobSyncAction) GetObject() interface{} {
	return a.job
}

func PipelineJobAddAction(job api.PipelineJob) *PipelineJobSyncAction {
	return NewPipelineJobSyncAction(job, SyncActionAdd)
}

func PipelineJobUpdateAction(job api.PipelineJob) *PipelineJobSyncAction {
	return NewPipelineJobSyncAction(job, SyncActionUpdate)
}

func PipelineJobDeleteAction(job api.PipelineJob) *PipelineJobSyncAction {
	return NewPipelineJobSyncAction(job, SyncActionDelete)
}

func NewPipelineJobSyncAction(job api.PipelineJob, actionType SyncActionType) *PipelineJobSyncAction {
	return &PipelineJobSyncAction{
		action: actionType,
		job:    job,
	}
}
