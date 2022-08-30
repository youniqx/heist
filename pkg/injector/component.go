package injector

import (
	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type component struct {
	Log    logr.Logger
	Config *Config
}

type Config struct {
	AgentImage string
	OpenShift  bool
}

func Component(config *Config) operator.Component {
	if config == nil {
		config = &Config{
			AgentImage: "youniqx/heist:latest",
		}
	}

	return &component{
		Log:    controllerruntime.Log.WithName("setup-agent-webhook"),
		Config: config,
	}
}

func (c *component) Register(api vault.API, mgr manager.Manager) error {
	handler := &Handler{
		Log:           controllerruntime.Log.WithName("agent-injector-webhook"),
		VaultAPI:      api,
		K8sClient:     mgr.GetClient(),
		Filter:        operator.NewFilter(),
		VaultAddress:  api.GetAddress(),
		AuthMountPath: managed.KubernetesAuthPath,
		Config:        c.Config,
	}

	mgr.GetWebhookServer().Register("/mutate-pod-agent-injector", handler)
	return nil
}
