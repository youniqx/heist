package agent

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func (a *agent) Stop() {
	a.StopChannel <- true
}

const watchRetryDelay = 5 * time.Second

//nolint:cyclop,gocognit
func (a *agent) Run() {
	var resultChannel <-chan watch.Event
	var informer watch.Interface
	var err error

	timer := time.NewTicker(watchRetryDelay)
	defer timer.Stop()

	fallbackResultChannel := make(<-chan watch.Event)

	resultChannel = fallbackResultChannel

	var configSyncTime time.Time

	for {
		select {
		case <-timer.C:
			if informer == nil {
				informer, err = a.ClientSet.HeistV1alpha1().VaultClientConfigs(a.Namespace).Watch(context.Background(), metav1.ListOptions{})
				if err != nil {
					a.Log.Info("failed to create informer for the targeted VaultClientConfig object", "error", err)
					resultChannel = fallbackResultChannel
				} else {
					resultChannel = informer.ResultChan()
				}
			}

			if configSyncTime.Before(time.Now().Add(-time.Minute)) {
				config, err := a.ClientSet.HeistV1alpha1().VaultClientConfigs(a.Namespace).Get(context.TODO(), a.Name, metav1.GetOptions{})
				if err != nil {
					a.Log.Info("failed to fetch VaultClientConfig object", "error", err)
					continue
				}
				a.updateConfig(config)
				configSyncTime = time.Now()
			}
		case <-a.StopChannel:
			if informer != nil {
				informer.Stop()
			}
			a.Status = &SyncStatus{
				Status: StatusStopped,
				Reason: "",
			}
			return
		case event, ok := <-resultChannel:
			if !ok {
				informer.Stop()
				informer = nil
				resultChannel = fallbackResultChannel
				continue
			}

			log := a.Log.WithValues("type", event.Type)

			var conf *v1alpha1.VaultClientConfig
			switch obj := event.Object.(type) {
			case *v1alpha1.VaultClientConfig:
				conf = obj
				log = log.WithValues("namespace", conf.Namespace, "name", conf.Name)
			default:
				log.Info("Received watch event for some unknown object. Ignoring...")
				continue
			}

			if conf.Namespace != a.Namespace || conf.Name != a.Name {
				continue
			}

			log = log.WithValues("object", conf)
			log.Info("Received watch event for VaultClientConfig")

			switch event.Type {
			case watch.Added, watch.Modified:
				a.updateConfig(conf)
			case watch.Bookmark:
			case watch.Deleted:
				a.updateConfig(conf)
			case watch.Error:
			}
		}
	}
}

func (a *agent) updateConfig(conf *v1alpha1.VaultClientConfig) {
	a.ConfigLock.Lock()
	defer a.ConfigLock.Unlock()

	newConfig := &config{
		API:          nil,
		ClientConfig: conf,
	}

	log := newConfig.AddToLogger(a.Log).WithValues("current_sync_state", a.Status.Status)

	if a.Config != nil && a.Config.ClientConfig != nil && reflect.DeepEqual(a.Config.ClientConfig.Spec, newConfig.ClientConfig.Spec) {
		log.Info("Config has not changed, skipping config update...")
		return
	}

	if newConfig.SameVault(a.Config) {
		newConfig.API = a.Config.API
		newConfig.Cache = a.Config.Cache
		log.Info("reusing Vault API instance from old config since they refer to the same Vault instance")
	} else {
		var err error

		cas := make([]core.StringSource, 0, len(newConfig.ClientConfig.Spec.CACerts))
		for _, cert := range newConfig.ClientConfig.Spec.CACerts {
			cas = append(cas, core.Value(cert))
		}

		if a.VaultToken != "" {
			newConfig.API, err = vault.NewAPI().
				WithAddressFrom(core.Value(newConfig.ClientConfig.Spec.Address)).
				WithTokenFrom(core.Value(a.VaultToken)).
				WithCAsFrom(cas...).
				Complete()
		} else {
			newConfig.API, err = vault.NewAPI().
				WithAddressFrom(core.Value(newConfig.ClientConfig.Spec.Address)).
				WithAuthProvider(kubernetesauth.AuthProvider(
					core.MountPath(newConfig.ClientConfig.Spec.AuthMountPath),
					core.Value(newConfig.ClientConfig.Spec.Role),
					core.File(a.TokenPath),
				)).
				WithCAsFrom(cas...).
				Complete()
		}
		if err != nil {
			a.Status = &SyncStatus{
				Status: StatusError,
				Reason: fmt.Sprintf("failed to create vault api instance: %v", err),
			}
			log.Info("failed to create new Vault API instance.", "error", err, "new_sync_state", a.Status.Status)
			return
		}
		newConfig.Cache = newCache(newConfig.API)
		log.Info("created new Vault API instance")
	}

	a.Status = &SyncStatus{
		Status: StatusSynced,
		Reason: "",
	}

	a.Config = newConfig
	log.Info("config synced successfully", "new_sync_state", a.Status.Status)

	for index, channel := range a.UpdateChannels {
		select {
		case channel <- true:
		default:
			log.Info("dropped message for update channel", "index", index)
		}
	}
}

func (a *agent) getConfig() *config {
	a.ConfigLock.Lock()
	defer a.ConfigLock.Unlock()
	return a.Config
}
