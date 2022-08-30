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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/client/heist.youniqx.com/v1alpha1/clientset/heist"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// For transparent login to common Cloud Platform e.g. GCP, AWS, Azure etc.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// agentCmd represents the agent command.
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Start the heist agent",
	ValidArgs: []string{
		"--address",
		"--client-config-name",
		"--client-config-namespace",
		"--kubernetes-config-path",
		"--kubernetes-master-url",
		"--secret-base-path",
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.PersistentFlags().String("kubernetes-master-url", defaultConfig.Agent.KubernetesMasterURL, "URL of the Kubernetes API server.")
	_ = viper.BindPFlag("agent.kubernetes_master_url", agentCmd.PersistentFlags().Lookup("kubernetes-master-url"))
	_ = agentCmd.RegisterFlagCompletionFunc("kubernetes-master-url", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	agentCmd.PersistentFlags().String("kubernetes-config-path", defaultConfig.Agent.KubernetesConfigPath, "Path to the Kubernetes config file.")
	_ = viper.BindPFlag("agent.kubernetes_config_path", agentCmd.PersistentFlags().Lookup("kubernetes-config-path"))
	_ = agentCmd.RegisterFlagCompletionFunc("kubernetes-config-path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})

	agentCmd.PersistentFlags().String("client-config-namespace", defaultConfig.Agent.ClientConfigNamespace, "Namespace containing the client config to watch.")
	_ = viper.BindPFlag("agent.client_config_namespace", agentCmd.PersistentFlags().Lookup("client-config-namespace"))
	_ = agentCmd.RegisterFlagCompletionFunc("client-config-namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

	agentCmd.PersistentFlags().String("client-config-name", defaultConfig.Agent.ClientConfigName, "Name of the client config object to watch.")
	_ = viper.BindPFlag("agent.client_config_name", agentCmd.PersistentFlags().Lookup("client-config-name"))
	_ = agentCmd.RegisterFlagCompletionFunc("client-config-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := createHeistClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		namespace := cmd.Flags().Lookup("client-config-namespace").Value.String()

		configs, err := clientset.HeistV1alpha1().VaultClientConfigs(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		names := make([]string, 0, len(configs.Items))
		for _, item := range configs.Items {
			names = append(names, item.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	})

	agentCmd.PersistentFlags().String("address", defaultConfig.Agent.Address, "Address the agent will be listening on.")
	_ = viper.BindPFlag("agent.address", agentCmd.PersistentFlags().Lookup("address"))
	_ = agentCmd.RegisterFlagCompletionFunc("address", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	agentCmd.PersistentFlags().String("secret-base-path", defaultConfig.Agent.SecretBasePath, "Base path for secrets synced by the agent.")
	_ = viper.BindPFlag("agent.secret_base_path", agentCmd.PersistentFlags().Lookup("secret-base-path"))
	_ = agentCmd.RegisterFlagCompletionFunc("secret-base-path", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})
}

func createHeistClientSet() (*heist.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}

	clientSet, err := heist.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

func createKubernetesClientSet() (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}
