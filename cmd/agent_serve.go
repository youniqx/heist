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
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/youniqx/heist/pkg/agent"
	"github.com/youniqx/heist/pkg/agentserver"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// serveCmd represents the preload command.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the agent server and serve the Agent API at the specified port",
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

		server := agentserver.New(instance)
		defer server.Stop()

		errChan := make(chan error)

		log := controllerruntime.Log.
			WithName("server").
			WithValues("address", heistConfig.Agent.Address)

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

			signal := <-c

			log.Info("Received Signal", "signal", signal)

			errChan <- fmt.Errorf("received signal: %v", signal)
		}()

		go func() {
			log.Info("starting server")
			errChan <- server.ListenAndServer(heistConfig.Agent.Address)
		}()

		err = <-errChan

		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server has gracefully shut down")
		} else {
			log.Info("server stopped because of an unexpected error", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	agentCmd.AddCommand(serveCmd)
}
