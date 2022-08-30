package v1alpha1

import (
	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type component struct {
	Log logr.Logger
}

func Component() operator.Component {
	return &component{
		Log: controllerruntime.Log.WithName("setup-webhook"),
	}
}

func (c *component) Register(api vault.API, mgr manager.Manager) error {
	if err := (&VaultKVSecret{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultKVSecret")
		return err
	}
	if err := (&VaultKVSecretEngine{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultKVSecretEngine")
		return err
	}
	if err := (&VaultBinding{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultBinding")
		return err
	}
	if err := (&VaultCertificateAuthority{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultCertificateAuthority")
		return err
	}
	if err := (&VaultSyncSecret{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultSyncSecret")
		return err
	}
	if err := (&VaultTransitEngine{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultTransitEngine")
		return err
	}
	if err := (&VaultTransitKey{}).SetupWebhookWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create webhook", "webhook", "VaultTransitKey")
		return err
	}
	// +kubebuilder:scaffold:webhook
	return nil
}
