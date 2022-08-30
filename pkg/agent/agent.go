package agent

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist"
	"github.com/youniqx/heist/pkg/erx"
	"github.com/youniqx/heist/pkg/vault"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type Agent interface {
	ListSecrets() (names []string, err error)
	FetchSecret(name string) (secret *Secret, err error)
	GetClientSecret() *Secret
	GetStatus() *SyncStatus
	Stop()
	CreateUpdateChannel(chan bool)
}

type agent struct {
	Log            logr.Logger
	ClientSet      *heist.Clientset
	Namespace      string
	Name           string
	Status         *SyncStatus
	Config         *config
	BasePath       string
	ConfigLock     sync.Mutex
	StopChannel    chan bool
	TokenPath      string
	VaultToken     string
	UpdateChannels []chan bool
}

func (a *agent) CreateUpdateChannel(updateChannel chan bool) {
	a.UpdateChannels = append(a.UpdateChannels, updateChannel)
}

// ErrInitAgentFailed is returned when the agent can't be initialized.
var ErrInitAgentFailed = erx.New("Heist Agent", "failed to initialize new agent")

var defaultLogger = controllerruntime.Log.WithName("agent")

func New(opts ...Option) (Agent, error) {
	instance := &agent{
		Log: defaultLogger,
		Status: &SyncStatus{
			Status: StatusNotYetSynced,
			Reason: "waiting for the first config sync to complete",
		},
		StopChannel: make(chan bool),
		TokenPath:   vault.DefaultKubernetesTokenPath,
		BasePath:    "/heist",
	}

	for _, opt := range opts {
		if err := opt(instance); err != nil {
			return nil, ErrInitAgentFailed.
				WithDetails("failed to evaluate one of the supplied options").
				WithCause(err)
		}
	}

	if instance.ClientSet == nil {
		return nil, ErrInitAgentFailed.WithDetails("no rest config or kubernetes credentials supplied")
	}

	if instance.Namespace == "" {
		return nil, ErrInitAgentFailed.WithDetails("namespace of the client config object is not set")
	}

	if instance.Name == "" {
		return nil, ErrInitAgentFailed.WithDetails("name of the client config object is not set")
	}

	instance.Log = instance.Log.WithValues("config_namespace", instance.Namespace, "config_name", instance.Name)

	go instance.Run()

	return instance, nil
}
