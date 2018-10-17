package sync

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

type PipelineSyncAction struct {
	action   SyncActionType
	pipeline *api.Pipeline
}

func (a *PipelineSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *PipelineSyncAction) GetObject() interface{} {
	return a.pipeline
}

func PipelineAddAction(pipeline *api.Pipeline) *PipelineSyncAction {
	return NewPipelineSyncAction(pipeline, SyncActionAdd)
}

func PipelineUpdateAction(pipeline *api.Pipeline) *PipelineSyncAction {
	return NewPipelineSyncAction(pipeline, SyncActionUpdate)
}

func PipelineDeleteAction(pipeline *api.Pipeline) *PipelineSyncAction {
	return NewPipelineSyncAction(pipeline, SyncActionDelete)
}

func NewPipelineSyncAction(pipeline *api.Pipeline, actionType SyncActionType) *PipelineSyncAction {
	return &PipelineSyncAction{
		action:   actionType,
		pipeline: pipeline,
	}
}
