package controllers

import (
	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/controllers/vaultbinding"
	"github.com/youniqx/heist/pkg/controllers/vaultcertificateauthority"
	"github.com/youniqx/heist/pkg/controllers/vaultcertificaterole"
	"github.com/youniqx/heist/pkg/controllers/vaultclientconfig"
	"github.com/youniqx/heist/pkg/controllers/vaultkvsecret"
	"github.com/youniqx/heist/pkg/controllers/vaultkvsecretengine"
	"github.com/youniqx/heist/pkg/controllers/vaultsyncsecret"
	"github.com/youniqx/heist/pkg/controllers/vaulttransitengine"
	"github.com/youniqx/heist/pkg/controllers/vaulttransitkey"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type component struct {
	Log                          logr.Logger
	SyncSecretNamespaceAllowList []string
}

type Config struct {
	SyncSecretNamespaceAllowList []string
}

func Component(config *Config) operator.Component {
	return &component{
		Log:                          controllerruntime.Log.WithName("setup-controller"),
		SyncSecretNamespaceAllowList: config.SyncSecretNamespaceAllowList,
	}
}

func (c *component) Register(api vault.API, mgr manager.Manager) error {
	filter := operator.NewFilter()

	if err := (&vaultkvsecret.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultKVSecret"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaultkvsecret-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultKVSecret")
		return err
	}
	if err := (&vaultkvsecretengine.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultKVSecretEngine"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaultkvsecret-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultKVSecretEngine")
		return err
	}
	if err := (&vaultbinding.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultBinding"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaultbinding-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultBinding")
		return err
	}
	if err := (&vaultcertificateauthority.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultCertificateAuthority"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaultcertificateauthority-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultCertificateAuthority")
		return err
	}
	if err := (&vaultcertificaterole.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultCertificateRole"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaultcertificaterole-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultCertificateRole")
		return err
	}
	if err := (&vaultclientconfig.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultClientConfig"),
		Scheme:      mgr.GetScheme(),
		Recorder:    mgr.GetEventRecorderFor("vaultclientconfig-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultClientConfig")
		return err
	}
	if err := (&vaultsyncsecret.Reconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		Recorder:           mgr.GetEventRecorderFor("vaultsyncsecret-controller"),
		VaultAPI:           api,
		EventFilter:        filter,
		NamespaceAllowList: c.SyncSecretNamespaceAllowList,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultSyncSecret")
		return err
	}
	if err := (&vaulttransitengine.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultTransitEngine"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaulttransitengine-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultTransitEngine")
		return err
	}
	if err := (&vaulttransitkey.Reconciler{
		Client:      mgr.GetClient(),
		Log:         controllerruntime.Log.WithName("controllers").WithName("VaultTransitKey"),
		Scheme:      mgr.GetScheme(),
		VaultAPI:    api,
		Recorder:    mgr.GetEventRecorderFor("vaulttransitkey-controller"),
		EventFilter: filter,
	}).SetupWithManager(mgr); err != nil {
		c.Log.Error(err, "unable to create controller", "controller", "VaultTransitKey")
		return err
	}
	// +kubebuilder:scaffold:builder
	return nil
}
