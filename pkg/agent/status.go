package agent

type SyncStatus struct {
	Status StatusType
	Reason string
}

type StatusType string

const (
	// StatusNotYetSynced indicates that the agent hasn't yet fetched it's config.
	StatusNotYetSynced StatusType = "not_yet_synced"
	// StatusSynced indicates that the agent has successfully fetched the config
	// and all configured secrets.
	StatusSynced StatusType = "synced"
	// StatusStopped indicates that the agent has been stopped and will no longer sync secrets.
	StatusStopped StatusType = "stopped"
	// StatusError indicates that an error has occurred.
	StatusError StatusType = "error"
)

func (a *agent) GetStatus() *SyncStatus {
	return &SyncStatus{
		Status: a.Status.Status,
		Reason: a.Status.Reason,
	}
}
