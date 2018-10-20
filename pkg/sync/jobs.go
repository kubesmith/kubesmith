package sync

import batchv1 "k8s.io/api/batch/v1"

type JobSyncAction struct {
	action SyncActionType
	job    batchv1.Job
}

func (a *JobSyncAction) GetAction() SyncActionType {
	return a.action
}

func (a *JobSyncAction) GetObject() interface{} {
	return a.job
}

func JobAddAction(job batchv1.Job) *JobSyncAction {
	return NewJobSyncAction(job, SyncActionAdd)
}

func JobUpdateAction(job batchv1.Job) *JobSyncAction {
	return NewJobSyncAction(job, SyncActionUpdate)
}

func JobDeleteAction(job batchv1.Job) *JobSyncAction {
	return NewJobSyncAction(job, SyncActionDelete)
}

func NewJobSyncAction(job batchv1.Job, actionType SyncActionType) *JobSyncAction {
	return &JobSyncAction{
		action: actionType,
		job:    job,
	}
}
