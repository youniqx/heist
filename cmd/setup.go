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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// setupCmd represents the setup command.
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configures Vault for use with the Heist Operator",
}

func init() {
	rootCmd.AddCommand(setupCmd)

	_ = setupCmd.MarkFlagRequired("vault-token")

	setupCmd.PersistentFlags().String("vault-token", defaultConfig.Setup.VaultToken, "Token used to authenticate in Vault.")
	_ = viper.BindPFlag("setup.vault_token", setupCmd.PersistentFlags().Lookup("vault-token"))
	_ = setupCmd.RegisterFlagCompletionFunc("vault-token", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	_ = setupCmd.MarkFlagRequired("vault-token")

	setupCmd.PersistentFlags().String("heist-namespace", defaultConfig.Setup.HeistNamespace, "Namespace containing the heist deployment.")
	_ = viper.BindPFlag("setup.heist_namespace", setupCmd.PersistentFlags().Lookup("heist-namespace"))
	_ = setupCmd.RegisterFlagCompletionFunc("heist-namespace", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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

	setupCmd.PersistentFlags().String("heist-service-account", defaultConfig.Setup.HeistServiceAccount, "Name of the service account used by the heist operator.")
	_ = viper.BindPFlag("setup.heist_service_account", setupCmd.PersistentFlags().Lookup("heist-service-account"))
	_ = setupCmd.RegisterFlagCompletionFunc("heist-service-account", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		clientset, err := createKubernetesClientSet()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		namespace := cmd.PersistentFlags().Lookup("heist-namespace").Value.String()

		serviceAccounts, err := clientset.CoreV1().ServiceAccounts(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		names := make([]string, 0, len(serviceAccounts.Items))
		for _, item := range serviceAccounts.Items {
			names = append(names, item.Name)
		}

		return names, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().StringSlice("vault-ca-cert", defaultConfig.Setup.VaultCACerts, "CA certs to verify Vault server certificate.")
	_ = viper.BindPFlag("setup.vault_ca_certs", setupCmd.PersistentFlags().Lookup("vault-ca-cert"))

	setupCmd.PersistentFlags().String("vault-scheme", defaultConfig.Setup.VaultScheme, "Scheme used to connect to vault (http or https)")
	_ = viper.BindPFlag("setup.vault_scheme", setupCmd.PersistentFlags().Lookup("vault-scheme"))

	setupCmd.PersistentFlags().String("heist-role-name", defaultConfig.Setup.RoleName, "Name of the role heist uses to authenticate in Vault.")
	_ = viper.BindPFlag("setup.role_name", setupCmd.PersistentFlags().Lookup("heist-role-name"))
	_ = setupCmd.RegisterFlagCompletionFunc("heist-role-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().String("heist-policy-name", defaultConfig.Setup.PolicyName, "Name of the policy containing ACL roles for the heist operator.")
	_ = viper.BindPFlag("setup.policy_name", setupCmd.PersistentFlags().Lookup("heist-policy-name"))
	_ = setupCmd.RegisterFlagCompletionFunc("heist-policy-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().String("kubernetes-host", defaultConfig.Setup.KubernetesHost, "Kubernetes API Server Host.")
	_ = viper.BindPFlag("setup.kubernetes_host", setupCmd.PersistentFlags().Lookup("kubernetes-host"))
	_ = setupCmd.RegisterFlagCompletionFunc("heist-policy-name", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().String("kubernetes-jwt-issuer", defaultConfig.Setup.KubernetesJWTIssuer, "Issuer of service account JWTs in the Kubernetes cluster.")
	_ = viper.BindPFlag("setup.kubernetes_jwt_issuer", setupCmd.PersistentFlags().Lookup("kubernetes-jwt-issuer"))
	_ = setupCmd.RegisterFlagCompletionFunc("kubernetes-jwt-issuer", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().String("kubernetes-jwt-ca-cert", defaultConfig.Setup.KubernetesJWTCACert, "CA certificate used to validate service account JWTs.")
	_ = viper.BindPFlag("setup.kubernetes_jwt_ca_cert", setupCmd.PersistentFlags().Lookup("kubernetes-jwt-ca-cert"))
	_ = setupCmd.RegisterFlagCompletionFunc("kubernetes-jwt-ca-cert", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})

	setupCmd.PersistentFlags().StringSlice("kubernetes-jwt-pem-key", defaultConfig.Setup.KubernetesJWTPemKeys, "One or more keys in PEM format used to validate service account JWTs.")
	_ = viper.BindPFlag("setup.kubernetes_jwt_pem_keys", setupCmd.PersistentFlags().Lookup("kubernetes-jwt-pem-key"))
	_ = setupCmd.RegisterFlagCompletionFunc("kubernetes-jwt-pem-key", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
}
