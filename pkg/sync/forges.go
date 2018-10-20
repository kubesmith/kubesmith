package sync

import (
	api "github.com/kubesmith/kubesmith/pkg/apis/kubesmith/v1"
)

type ForgeSyncAction struct {
	action SyncActionType
	forge  api.Forge
}

func (a *ForgeSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *ForgeSyncAction) GetObject() interface{} {
	return a.forge
}

func ForgeAddAction(forge api.Forge) *ForgeSyncAction {
	return NewForgeSyncAction(forge, SyncActionAdd)
}

func ForgeUpdateAction(forge api.Forge) *ForgeSyncAction {
	return NewForgeSyncAction(forge, SyncActionUpdate)
}

func ForgeDeleteAction(forge api.Forge) *ForgeSyncAction {
	return NewForgeSyncAction(forge, SyncActionDelete)
}

func NewForgeSyncAction(forge api.Forge, actionType SyncActionType) *ForgeSyncAction {
	return &ForgeSyncAction{
		action: actionType,
		forge:  forge,
	}
}
