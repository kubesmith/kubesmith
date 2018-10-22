package v1

const (
	DefaultNamespace = "kubesmith"
)

const (
	PhaseEmpty     = ""
	PhaseQueued    = "Queued"
	PhaseRunning   = "Running"
	PhaseSucceeded = "Succeeded"
	PhaseFailed    = "Failed"
)

type Phase string
