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
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// k8sCmd represents the k8s command.
var k8sCmd = &cobra.Command{
	Use:   "k8s",
	Short: "Configures an in-cluster Vault instance for use with the Heist Operator",
	ValidArgs: []string{
		"--heist-namespace",
		"--heist-policy-name",
		"--heist-role-name",
		"--heist-service-account",
		"--kubernetes-host",
		"--kubernetes-jwt-ca-cert",
		"--kubernetes-jwt-issuer",
		"--kubernetes-jwt-pem-key",
		"--vault-namespace",
		"--vault-port",
		"--vault-service",
		"--vault-token",
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		cobra.CheckErr(err)

		if cfgPath, ok := os.LookupEnv("KUBECONFIG"); ok {
			config, err = clientcmd.BuildConfigFromFlags("", cfgPath)
			cobra.CheckErr(err)
		}

		heistConfig := &HeistConfig{}
		cobra.CheckErr(viper.Unmarshal(heistConfig))

		err = operator.SetupOperator(&operator.SetupConfig{
			VaultNamespace:       heistConfig.Setup.VaultNamespace,
			VaultServiceName:     heistConfig.Setup.VaultServiceName,
			VaultPort:            heistConfig.Setup.VaultPort,
			VaultToken:           heistConfig.Setup.VaultToken,
			VaultCAs:             parseValues(heistConfig.Setup.VaultCACerts),
			VaultScheme:          heistConfig.Setup.VaultScheme,
			VaultURL:             "",
			KubernetesHost:       heistConfig.Setup.KubernetesHost,
			KubernetesJWTIssuer:  heistConfig.Setup.KubernetesJWTIssuer,
			KubernetesJWTCACert:  heistConfig.Setup.KubernetesJWTCACert,
			KubernetesJWTPemKeys: heistConfig.Setup.KubernetesJWTPemKeys,
			PolicyName:           heistConfig.Setup.PolicyName,
			RoleName:             heistConfig.Setup.RoleName,
			HeistNamespace:       heistConfig.Setup.HeistNamespace,
			HeistServiceAccount:  heistConfig.Setup.HeistServiceAccount,
			RESTConfig:           config,
		})
		cobra.CheckErr(err)
	},
}

func init() {
	setupCmd.AddCommand(k8sCmd)

	k8sCmd.Flags().String("vault-namespace", defaultConfig.Setup.VaultNamespace, "Namespace containing Vault instance.")
	_ = viper.BindPFlag("setup.vault_namespace", k8sCmd.Flags().Lookup("vault-namespace"))
	_ = k8sCmd.RegisterFlagCompletionFunc("vault-namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
	_ = k8sCmd.MarkFlagRequired("vault-namespace")

	k8sCmd.Flags().String("vault-service", defaultConfig.Setup.VaultServiceName, "Name of the Vault service.")
	_ = viper.BindPFlag("setup.vault_service_name", k8sCmd.Flags().Lookup("vault-service"))
	_ = k8sCmd.RegisterFlagCompletionFunc("vault-service", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := createKubernetesClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		namespace := cmd.Flags().Lookup("vault-namespace").Value.String()

		services, err := clientset.CoreV1().Services(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		names := make([]string, 0, len(services.Items))
		for _, item := range services.Items {
			names = append(names, item.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	})
	_ = k8sCmd.MarkFlagRequired("vault-service")

	k8sCmd.Flags().String("vault-port", defaultConfig.Setup.VaultPort, "Port the Vault service listens on.")
	_ = viper.BindPFlag("setup.vault_port", k8sCmd.Flags().Lookup("vault-port"))
	_ = k8sCmd.RegisterFlagCompletionFunc("vault-port", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := createKubernetesClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		namespace := cmd.Flags().Lookup("vault-namespace").Value.String()
		name := cmd.Flags().Lookup("vault-service").Value.String()

		service, err := clientset.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		var ports []string

		for _, port := range service.Spec.Ports {
			ports = append(ports, strconv.Itoa(int(port.Port)))
		}

		return ports, cobra.ShellCompDirectiveNoFileComp
	})
	_ = k8sCmd.MarkFlagRequired("vault-port")
}
