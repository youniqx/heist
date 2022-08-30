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
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault"
	"gopkg.in/yaml.v3"
)

var (
	cfgFile   string
	version   string
	commit    string
	buildTime string
	tag       = "latest"
)

func formatVersionString() string {
	if commit == "" {
		return "local build"
	}

	return fmt.Sprintf("\u001B[32m%s\u001B[0m (\u001B[32m%s\u001B[0m @ \u001B[32m%s\u001B[0m)", version, buildTime, commit)
}

var rootCmd = &cobra.Command{
	Use:     "heist",
	Short:   "A tool chain to securely automate management of secrets in vault.",
	Version: formatVersionString(),
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

//nolint:gomnd
var defaultConfig = &HeistConfig{
	Vault: &VaultConfig{
		Address:                 "",
		Role:                    "",
		Token:                   "",
		KubernetesAuthMountPath: managed.KubernetesAuthPath,
		JWTPath:                 vault.DefaultKubernetesTokenPath,
	},
	Operator: &OperatorConfig{
		MetricsBindAddress:           ":8080",
		HealthProbeBindAddress:       ":8081",
		WebhookPort:                  9443,
		LeaderElect:                  false,
		LeaderElectionID:             "8b2618e3.youniqx.com",
		AgentImage:                   fmt.Sprintf("youniqx/heist:%s", tag),
		SyncSecretNamespaceAllowList: nil,
	},
	Agent: &AgentConfig{
		KubernetesMasterURL:   "",
		KubernetesConfigPath:  "",
		ClientConfigNamespace: "",
		ClientConfigName:      "",
		SecretBasePath:        "/heist",
		Address:               ":8080",
	},
	Setup: &SetupConfig{
		VaultNamespace:       "",
		VaultServiceName:     "",
		VaultPort:            "",
		VaultCACerts:         nil,
		VaultScheme:          "http",
		VaultToken:           "",
		VaultURL:             "",
		PolicyName:           "heist",
		RoleName:             "heist",
		HeistNamespace:       "heist-system",
		HeistServiceAccount:  "heist",
		KubernetesHost:       "https://kubernetes.default.svc.cluster.local",
		KubernetesJWTIssuer:  "",
		KubernetesJWTCACert:  "",
		KubernetesJWTPemKeys: nil,
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.heist.yaml)")

	loadDefaultConfig(defaultConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigName(".heist")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

type HeistConfig struct {
	Vault    *VaultConfig    `mapstructure:"vault" yaml:"vault" json:"vault"`
	Agent    *AgentConfig    `mapstructure:"agent" yaml:"agent" json:"agent"`
	Operator *OperatorConfig `mapstructure:"operator" yaml:"operator" json:"operator"`
	Setup    *SetupConfig    `mapstructure:"setup" yaml:"setup" json:"setup"`
}

type VaultConfig struct {
	Address                 string   `mapstructure:"address" yaml:"address" json:"address"`
	CACerts                 []string `mapstructure:"ca_certs" yaml:"ca_certs" json:"ca_certs"`
	Role                    string   `mapstructure:"role" yaml:"role" json:"role"`
	Token                   string   `mapstructure:"token" yaml:"token" json:"token"`
	KubernetesAuthMountPath string   `mapstructure:"kubernetes_auth_mount_path" yaml:"kubernetes_auth_mount_path" json:"kubernetes_auth_mount_path"`
	JWTPath                 string   `mapstructure:"jwt_path" yaml:"jwt_path" json:"jwt_path"`
}

type AgentConfig struct {
	KubernetesMasterURL   string `mapstructure:"kubernetes_master_url" yaml:"kubernetes_master_url" json:"kubernetes_master_url"`
	KubernetesConfigPath  string `mapstructure:"kubernetes_config_path" yaml:"kubernetes_config_path" json:"kubernetes_config_path"`
	ClientConfigNamespace string `mapstructure:"client_config_namespace" yaml:"client_config_namespace" json:"client_config_namespace"`
	ClientConfigName      string `mapstructure:"client_config_name" yaml:"client_config_name" json:"client_config_name"`
	SecretBasePath        string `mapstructure:"secret_base_path" yaml:"secret_base_path" json:"secret_base_path"`
	Address               string `mapstructure:"address" yaml:"address" json:"address"`
}

type SetupConfig struct {
	VaultNamespace       string   `mapstructure:"vault_namespace" yaml:"vault_namespace" json:"vault_namespace"`
	VaultServiceName     string   `mapstructure:"vault_service_name" yaml:"vault_service_name" json:"vault_service_name"`
	VaultPort            string   `mapstructure:"vault_port" yaml:"vault_port" json:"vault_port"`
	VaultCACerts         []string `mapstructure:"vault_ca_certs" yaml:"vault_ca_certs" json:"vault_ca_certs"`
	VaultScheme          string   `mapstructure:"vault_scheme" yaml:"vault_scheme" json:"vault_scheme"`
	VaultToken           string   `mapstructure:"vault_token" yaml:"vault_token" json:"vault_token"`
	VaultURL             string   `mapstructure:"vault_url" yaml:"vault_url" json:"vault_url"`
	PolicyName           string   `mapstructure:"policy_name" yaml:"policy_name" json:"policy_name"`
	RoleName             string   `mapstructure:"role_name" yaml:"role_name" json:"role_name"`
	HeistNamespace       string   `mapstructure:"heist_namespace" yaml:"heist_namespace" json:"heist_namespace"`
	HeistServiceAccount  string   `mapstructure:"heist_service_account" yaml:"heist_service_account" json:"heist_service_account"`
	KubernetesHost       string   `mapstructure:"kubernetes_host" yaml:"kubernetes_host" json:"kubernetes_host"`
	KubernetesJWTIssuer  string   `mapstructure:"kubernetes_jwt_issuer" yaml:"kubernetes_jwt_issuer" json:"kubernetes_jwt_issuer"`
	KubernetesJWTCACert  string   `mapstructure:"kubernetes_jwt_ca_cert" yaml:"kubernetes_jwt_ca_cert" json:"kubernetes_jwt_ca_cert"`
	KubernetesJWTPemKeys []string `mapstructure:"kubernetes_jwt_pem_keys" yaml:"kubernetes_jwt_pem_keys" json:"kubernetes_jwt_pem_keys"`
}

type OperatorConfig struct {
	MetricsBindAddress           string   `mapstructure:"metrics_bind_address" yaml:"metrics_bind_address" json:"metrics_bind_address"`
	HealthProbeBindAddress       string   `mapstructure:"health_probe_bind_address" yaml:"health_probe_bind" json:"health_probe_bind"`
	WebhookPort                  int      `mapstructure:"webhook_port" yaml:"webhook_port" json:"webhook_port"`
	LeaderElect                  bool     `mapstructure:"leader_elect" yaml:"leader_elect" json:"leader_elect"`
	LeaderElectionID             string   `mapstructure:"leader_election_id" yaml:"leader_election_id" json:"leader_election_id"`
	AgentImage                   string   `mapstructure:"agent_image" yaml:"agent_image" json:"agent_image"`
	SyncSecretNamespaceAllowList []string `mapstructure:"sync_secret_namespace_allow_list" yaml:"sync_secret_namespace_allow_list" json:"sync_secret_namespace_allow_list"`
}

func loadDefaultConfig(value interface{}) {
	buffer := encodeToBuffer(value)
	viper.SetConfigType("yaml")
	cobra.CheckErr(viper.MergeConfig(&buffer))
}

func encodeToBuffer(value interface{}) bytes.Buffer {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	defer encoder.Close()
	cobra.CheckErr(encoder.Encode(value))
	return buffer
}

func parseValue(value string) string {
	switch {
	case strings.HasPrefix(value, "file:"):
		data, err := os.ReadFile(strings.TrimPrefix(value, "file:"))
		cobra.CheckErr(err)
		return string(data)
	case strings.HasPrefix(value, "env:"):
		return os.Getenv(strings.TrimPrefix(value, "file:"))
	default:
		return value
	}
}

func parseValues(values []string) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = parseValue(value)
	}
	return result
}
