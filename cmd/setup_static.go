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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/operator"
)

// staticCmd represents the static command.
var staticCmd = &cobra.Command{
	Use:   "static",
	Short: "Configures a Vault instance for use with the Heist Operator",
	ValidArgs: []string{
		"--heist-namespace",
		"--heist-policy-name",
		"--heist-role-name",
		"--heist-service-account",
		"--kubernetes-host",
		"--kubernetes-jwt-ca-cert",
		"--kubernetes-jwt-issuer",
		"--kubernetes-jwt-pem-key",
		"--vault-token",
		"--vault-url",
	},
	Run: func(cmd *cobra.Command, args []string) {
		heistConfig := &HeistConfig{}
		cobra.CheckErr(viper.Unmarshal(heistConfig))

		err := operator.SetupOperator(&operator.SetupConfig{
			VaultNamespace:       "",
			VaultServiceName:     "",
			VaultPort:            "",
			VaultToken:           heistConfig.Setup.VaultToken,
			VaultURL:             heistConfig.Setup.VaultURL,
			VaultCAs:             parseValues(heistConfig.Setup.VaultCACerts),
			VaultScheme:          heistConfig.Setup.VaultScheme,
			KubernetesHost:       heistConfig.Setup.KubernetesHost,
			KubernetesJWTIssuer:  heistConfig.Setup.KubernetesJWTIssuer,
			KubernetesJWTCACert:  heistConfig.Setup.KubernetesJWTCACert,
			KubernetesJWTPemKeys: heistConfig.Setup.KubernetesJWTPemKeys,
			PolicyName:           heistConfig.Setup.PolicyName,
			RoleName:             heistConfig.Setup.RoleName,
			HeistNamespace:       heistConfig.Setup.HeistNamespace,
			HeistServiceAccount:  heistConfig.Setup.HeistServiceAccount,
			RESTConfig:           nil,
		})
		cobra.CheckErr(err)
	},
}

func init() {
	setupCmd.AddCommand(staticCmd)

	staticCmd.Flags().String("vault-url", defaultConfig.Setup.VaultToken, "URL to the Vault instance you want to configure.")
	_ = viper.BindPFlag("setup.vault_url", staticCmd.Flags().Lookup("vault-url"))
	_ = staticCmd.RegisterFlagCompletionFunc("vault-url", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	})
	_ = staticCmd.MarkFlagRequired("vault-url")
}
