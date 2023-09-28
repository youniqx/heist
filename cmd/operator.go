/*
Copyright 2022 youniqx Identity AG.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers"
	"github.com/youniqx/heist/pkg/injector"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

// controllerCmd represents the controller command.
var controllerCmd = &cobra.Command{
	Use:   "operator",
	Short: "Starts the heist operator",
	ValidArgs: []string{
		"--health-probe-bind-address",
		"--leader-elect",
		"--metrics-bind-address",
		"--vault-address",
		"--vault-jwt-path",
		"--vault-kubernetes-auth-mount-path",
		"--vault-role",
		"--vault-token",
		"--webhook-port",
	},
	Run: func(cmd *cobra.Command, args []string) {
		opts := zap.Options{
			Development: true,
		}

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

		heistConfig := &HeistConfig{}
		cobra.CheckErr(viper.Unmarshal(heistConfig))

		cas := make([]core.StringSource, 0, len(heistConfig.Vault.CACerts))
		for _, cert := range heistConfig.Vault.CACerts {
			cas = append(cas, core.File(cert))
		}

		var api vault.API
		var err error
		if heistConfig.Vault.Token != "" {
			api, err = vault.NewAPI().
				WithAddressFrom(core.Value(heistConfig.Vault.Address)).
				WithTokenFrom(core.Value(heistConfig.Vault.Token)).
				WithCAsFrom(cas...).
				Complete()
		} else {
			api, err = vault.NewAPI().
				WithAddressFrom(core.Value(heistConfig.Vault.Address)).
				WithCAsFrom(cas...).
				WithAuthProvider(kubernetesauth.AuthProvider(
					core.MountPath(heistConfig.Vault.KubernetesAuthMountPath),
					core.Value(heistConfig.Vault.Role),
					core.File(heistConfig.Vault.JWTPath),
				)).
				Complete()
		}

		if err != nil {
			setupLog.Error(err, "unable to create Vault API instance")
			os.Exit(1)
		}

		cfg := ctrl.GetConfigOrDie()

		openshift, err := isOpenshift(cfg)
		if err != nil {
			setupLog.Error(err, "unable to determine if running in openshift")
			os.Exit(1)
		}

		mgr, err := operator.Create().
			WithVaultAPI(api).
			WithRestConfig(cfg).
			WithOptions(generateManagerConfig(heistConfig)).
			Register(controllers.Component(&controllers.Config{
				SyncSecretNamespaceAllowList: heistConfig.Operator.SyncSecretNamespaceAllowList,
			})).
			Register(heistv1alpha1.Component()).
			Register(injector.Component(&injector.Config{
				AgentImage: heistConfig.Operator.AgentImage,
				OpenShift:  openshift,
			})).
			Complete()
		if err != nil {
			setupLog.Error(err, "unable to create manager")
			os.Exit(1)
		}

		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	},
}

func generateManagerConfig(heistConfig *HeistConfig) ctrl.Options {
	options := ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: heistConfig.Operator.MetricsBindAddress,
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: heistConfig.Operator.WebhookPort,
		}),
		HealthProbeBindAddress: heistConfig.Operator.HealthProbeBindAddress,
		LeaderElection:         heistConfig.Operator.LeaderElect,
		LeaderElectionID:       heistConfig.Operator.LeaderElectionID,
	}

	switch watchedNamespaces := os.Getenv("WATCH_NAMESPACE"); {
	case watchedNamespaces == "":
		setupLog.Info("Operator Scope: global")
		return options
	case strings.Contains(watchedNamespaces, ","):
		namespaces := strings.Split(watchedNamespaces, ",")
		setupLog.Info("Operator Scope: multi namespace", "namespaces", namespaces)

		options.NewCache = func(config *rest.Config, opts cache.Options) (cache.Cache, error) {
			opts.DefaultNamespaces = getNamespacedCacheConfig(namespaces)
			return cache.New(config, opts)
		}

		return options
	default:
		setupLog.Info("Operator Scope: single namespace", "namespaces", []string{watchedNamespaces})
		options.Cache.DefaultNamespaces = getNamespacedCacheConfig([]string{watchedNamespaces})
		return options
	}
}

func getNamespacedCacheConfig(namespaces []string) (namespaceConfig map[string]cache.Config) {
	namespaceConfig = map[string]cache.Config{}

	for _, namespace := range namespaces {
		// leave all cache.Config fields nil to use default settings
		namespaceConfig[namespace] = cache.Config{}
	}

	return namespaceConfig
}

func init() {
	rootCmd.AddCommand(controllerCmd)

	controllerCmd.Flags().String("vault-address", defaultConfig.Vault.Address, "Address of the Vault instance the operator manages.")
	_ = viper.BindPFlag("vault.address", controllerCmd.Flags().Lookup("vault-address"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("vault-role", defaultConfig.Vault.Role, "Role used by the operator to authenticate in the Vault instance when using Kubernetes Auth.")
	_ = viper.BindPFlag("vault.role", controllerCmd.Flags().Lookup("vault-role"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-role", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("vault-token", defaultConfig.Vault.Token, "Token used by the operator to authenticate in the Vault instance when using Token Auth.")
	_ = viper.BindPFlag("vault.token", controllerCmd.Flags().Lookup("vault-token"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-token", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("vault-kubernetes-auth-mount-path", defaultConfig.Vault.KubernetesAuthMountPath, "Path of the Kubernetes Auth Engine mounted in Vault used to authenticate in Vault.")
	_ = viper.BindPFlag("vault.kubernetes_auth_mount_path", controllerCmd.Flags().Lookup("vault-kubernetes-auth-mount-path"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-kubernetes-auth-mount-path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("vault-jwt-path", defaultConfig.Vault.JWTPath, "Path to the file containing the JWT used to authenticate in Vault when using Kubernetes Auth.")
	_ = viper.BindPFlag("vault.jwt_path", controllerCmd.Flags().Lookup("vault-jwt-path"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-jwt-path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})

	controllerCmd.Flags().String("metrics-bind-address", defaultConfig.Operator.MetricsBindAddress, "The address the metric endpoint binds to.")
	_ = viper.BindPFlag("operator.metrics_bind_address", controllerCmd.Flags().Lookup("metrics-bind-address"))
	_ = controllerCmd.RegisterFlagCompletionFunc("metrics-bind-address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().StringSlice("vault-ca-cert", defaultConfig.Vault.CACerts, "CA certs to verify Vault server certificate.")
	_ = viper.BindPFlag("vault.ca_certs", controllerCmd.Flags().Lookup("vault-ca-cert"))
	_ = controllerCmd.RegisterFlagCompletionFunc("vault-ca-cert", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("health-probe-bind-address", defaultConfig.Operator.HealthProbeBindAddress, "The address the probe endpoint binds to.")
	_ = viper.BindPFlag("operator.health_probe_bind_address", controllerCmd.Flags().Lookup("health-probe-bind-address"))
	_ = controllerCmd.RegisterFlagCompletionFunc("health-probe-bind-address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().Int("webhook-port", defaultConfig.Operator.WebhookPort, "The port the webhook server listens on.")
	_ = viper.BindPFlag("operator.webhook_port", controllerCmd.Flags().Lookup("webhook-port"))
	_ = controllerCmd.RegisterFlagCompletionFunc("webhook-port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().String("agent-image", defaultConfig.Operator.AgentImage, "The image that will be injected into pods as the heist agent.")
	_ = viper.BindPFlag("operator.agent_image", controllerCmd.Flags().Lookup("agent-image"))
	_ = controllerCmd.RegisterFlagCompletionFunc("agent-image", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	controllerCmd.Flags().Bool("leader-elect", defaultConfig.Operator.LeaderElect, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	_ = viper.BindPFlag("operator.leader_elect", controllerCmd.Flags().Lookup("leader-elect"))

	controllerCmd.Flags().StringSlice("sync-secret-namespace", defaultConfig.Operator.SyncSecretNamespaceAllowList, "Allow list of namespaces to which values can be synced")
	_ = viper.BindPFlag("operator.sync_secret_namespace_allow_list", controllerCmd.Flags().Lookup("sync-secret-namespace"))
	_ = controllerCmd.RegisterFlagCompletionFunc("sync-secret-namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := createKubernetesClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		names := make([]string, 0, len(namespaces.Items))
		for _, item := range namespaces.Items {
			names = append(names, item.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	})

	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(heistv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func isOpenshift(cfg *rest.Config) (bool, error) {
	const sccGroup, sccKind = "security.openshift.io", "SecurityContextConstraints"

	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}

	_, resourceLists, err := client.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}

	for _, rl := range resourceLists {
		if strings.HasPrefix(rl.GroupVersion, sccGroup+"/") {
			for _, r := range rl.APIResources {
				if r.Kind == sccKind {
					log.Println("detected OpenShift environment")
					return true, nil
				}
			}
		}
	}

	return false, nil
}
