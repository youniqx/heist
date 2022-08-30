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
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/agent"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	secretFolderPerm = 0o750
	secretFilePerm   = 0o640
)

// syncCmd represents the preload command.
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets once and then quit",
	Run: func(cmd *cobra.Command, args []string) {
		opts := zap.Options{
			Development: true,
		}

		controllerruntime.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

		heistConfig := &HeistConfig{}
		cobra.CheckErr(viper.Unmarshal(heistConfig))

		instance, err := agent.New(
			agent.WithKubeConfig(heistConfig.Agent.KubernetesMasterURL, heistConfig.Agent.KubernetesConfigPath),
			agent.WithClientConfig(heistConfig.Agent.ClientConfigNamespace, heistConfig.Agent.ClientConfigName),
			agent.WithBasePath(heistConfig.Agent.SecretBasePath),
		)
		cobra.CheckErr(err)
		defer instance.Stop()

		cobra.CheckErr(syncSecret(instance))
	},
}

func init() {
	agentCmd.AddCommand(syncCmd)
}

func syncSecret(instance agent.Agent) error {
	log := controllerruntime.Log.WithName("agent-sync")

	for instance.GetStatus().Status != agent.StatusSynced {
		log.Info("Waiting for config sync...")
		time.Sleep(time.Second)
	}
	log.Info("Config has been synced, now syncing secrets...")

	secrets, err := instance.ListSecrets()
	cobra.CheckErr(err)
	log.Info("Fetched list of secrets")

	for _, name := range secrets {
		secretLog := log.WithValues("secret_name", name)
		secret, err := instance.FetchSecret(name)
		if err != nil {
			secretLog.Info("Failed to fetch secret", "error", err)
			return err
		}

		secretLog = secretLog.WithValues("output_path", secret.OutputPath)

		if err := os.MkdirAll(filepath.Dir(secret.OutputPath), secretFolderPerm); err != nil {
			secretLog.Info("Failed to create output folder", "error", err)
			return err
		}

		if err := os.WriteFile(secret.OutputPath, []byte(secret.Value), secretFilePerm); err != nil {
			secretLog.Info("Failed to write secret to disk", "error", err)
			return err
		}

		secretLog.Info("Successfully wrote secret to disk")
	}

	return nil
}
