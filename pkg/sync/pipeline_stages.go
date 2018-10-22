package sync

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

type PipelineStageSyncAction struct {
	action SyncActionType
	stage  api.PipelineStage
}

func (a *PipelineStageSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *PipelineStageSyncAction) GetObject() interface{} {
	return a.stage
}

func PipelineStageAddAction(stage api.PipelineStage) *PipelineStageSyncAction {
	return NewPipelineStageSyncAction(stage, SyncActionAdd)
}

func PipelineStageUpdateAction(stage api.PipelineStage) *PipelineStageSyncAction {
	return NewPipelineStageSyncAction(stage, SyncActionUpdate)
}

func PipelineStageDeleteAction(stage api.PipelineStage) *PipelineStageSyncAction {
	return NewPipelineStageSyncAction(stage, SyncActionDelete)
}

func NewPipelineStageSyncAction(stage api.PipelineStage, actionType SyncActionType) *PipelineStageSyncAction {
	return &PipelineStageSyncAction{
		action: actionType,
		stage:  stage,
	}
}
