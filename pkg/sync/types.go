package sync

const (
	SyncActionAdd    = "add"
	SyncActionUpdate = "update"
	SyncActionDelete = "delete"
)

type SyncActionType string

type SyncAction interface {
	GetAction() SyncActionType
	GetObject() interface{}
}
