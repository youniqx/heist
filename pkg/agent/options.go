package agent

import (
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Option func(a *agent) error

func WithRestConfig(config *rest.Config) Option {
	return func(a *agent) (err error) {
		a.ClientSet, err = heist.NewForConfig(config)
		return err
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(a *agent) (err error) {
		a.Log = logger
		return err
	}
}

func WithTokenPath(path string) Option {
	return func(a *agent) (err error) {
		a.TokenPath = path
		return err
	}
}

func WithBasePath(path string) Option {
	return func(a *agent) (err error) {
		a.BasePath = path
		return err
	}
}

func WithVaultToken(token string) Option {
	return func(a *agent) (err error) {
		a.VaultToken = token
		return err
	}
}

func WithKubeConfig(masterURL string, kubeConfigPath string) Option {
	return func(a *agent) (err error) {
		config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeConfigPath)
		if err == nil {
			a.ClientSet, err = heist.NewForConfig(config)
			return err
		}

		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err == nil {
			a.ClientSet, err = heist.NewForConfig(config)
			return err
		}

		return err
	}
}

func WithClientConfig(namespace string, name string) Option {
	return func(a *agent) error {
		a.Namespace = namespace
		a.Name = name
		return nil
	}
}
