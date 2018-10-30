package sync

import (
	corev1 "k8s.io/api/core/v1"
)

type PodSyncAction struct {
	action SyncActionType
	pod    corev1.Pod
}

func (a *PodSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *PodSyncAction) GetObject() interface{} {
	return a.pod
}

func PodAddAction(pod corev1.Pod) *PodSyncAction {
	return NewPodSyncAction(pod, SyncActionAdd)
}

func PodUpdateAction(pod corev1.Pod) *PodSyncAction {
	return NewPodSyncAction(pod, SyncActionUpdate)
}

func PodDeleteAction(pod corev1.Pod) *PodSyncAction {
	return NewPodSyncAction(pod, SyncActionDelete)
}

func NewPodSyncAction(pod corev1.Pod, actionType SyncActionType) *PodSyncAction {
	return &PodSyncAction{
		action: actionType,
		pod:    pod,
	}
}
