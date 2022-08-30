package agentserver

import (
	"os"
	"path/filepath"
	"time"
)

const (
	pollInterval            = 5 * time.Second
	syncInterval            = 2 * time.Minute
	secretFolderPerm        = 0o755
	minimumAgentPermissions = 0o600
)

func (s *server) WorkerLoop(workChannel chan bool) {
	defer close(workChannel)

	s.GlobalWorkChannel = make(chan bool)
	defer close(s.GlobalWorkChannel)

	timer := time.NewTicker(pollInterval)
	defer timer.Stop()

	s.performSync()

	updateChannel := make(chan bool, 1)
	s.Agent.CreateUpdateChannel(updateChannel)
	defer close(updateChannel)
	for {
		select {
		case <-updateChannel:
			s.NextSyncTime = time.Now()
			s.Log.Info("syncing secrets after config update")
			s.performSync()
		case <-timer.C:
			s.performSync()
		case <-s.StopChannel:
			s.Log.Info("Received stop signal, shutting down agent & server")
			s.setSyncStatus(false)
			s.Agent.Stop()
			if err := s.StopListening(); err != nil {
				s.Log.Info("failed to shutdown server", "error", err)
			}
			s.Log.Info("Quitting work loop")
			return
		}
	}
}

func (s *server) performSync() {
	s.SyncLock.Lock()
	defer s.SyncLock.Unlock()
	if s.NextSyncTime.After(time.Now()) {
		return
	}
	s.Log.Info("started syncing secrets")
	if err := s.TrySyncingSecrets(); err != nil {
		s.Log.Info("Failed to sync secrets", "error", err)
		s.NextSyncTime = time.Now().Add(pollInterval)
		return
	}
	s.Log.Info("Successfully synced secrets to disk")
	s.NextSyncTime = time.Now().Add(syncInterval)
	s.setSyncStatus(true)
}

func (s *server) setSyncStatus(synced bool) {
	s.StatusLock.Lock()
	defer s.StatusLock.Unlock()
	s.SyncCompleted = synced
}

func (s *server) IsSynced() bool {
	s.StatusLock.Lock()
	defer s.StatusLock.Unlock()
	return s.SyncCompleted
}

//nolint:cyclop,gocognit
func (s *server) TrySyncingSecrets() error {
	clientConfig := s.Agent.GetClientSecret()
	if err := os.MkdirAll(filepath.Dir(clientConfig.OutputPath), secretFolderPerm); err != nil {
		return err
	}
	if err := os.WriteFile(clientConfig.OutputPath, []byte(clientConfig.Value), clientConfig.Mode|minimumAgentPermissions); err != nil {
		return err
	}

	secrets, err := s.Agent.ListSecrets()
	if err != nil {
		return err
	}

	secretNames := make([]string, 0, len(secrets))

	for _, name := range secrets {
		log := s.Log.WithValues("secret_name", name)

		secret, err := s.Agent.FetchSecret(name)
		if err != nil {
			log.Info("failed to fetch secret")
			return err
		}

		mode := secret.Mode | minimumAgentPermissions

		log = log.WithValues("output_path", secret.OutputPath, "permissions", mode)

		secretNames = append(secretNames, secret.Name)

		if err := os.MkdirAll(filepath.Dir(secret.OutputPath), secretFolderPerm); err != nil {
			return err
		}

		switch currentPath := s.SyncedSecretMap[secret.Name]; {
		case currentPath == "":
			log.Info("Secret is new, writing it to the disk for the first time")
			if err := os.WriteFile(secret.OutputPath, []byte(secret.Value), mode); err != nil {
				return err
			}
		case currentPath == secret.OutputPath:
			log.Info("Secret already exists on disk, updating it")
			if err := os.WriteFile(secret.OutputPath, []byte(secret.Value), mode); err != nil {
				return err
			}
		default:
			log.Info("Secret output path has changed removing file at old location and writing secret to new location")
			if err := os.Remove(currentPath); err != nil {
				return err
			}
			if err := os.WriteFile(secret.OutputPath, []byte(secret.Value), mode); err != nil {
				return err
			}
		}

		s.SyncedSecretMap[secret.Name] = secret.OutputPath
	}

	for name, path := range s.SyncedSecretMap {
		log := s.Log.WithValues("secret_name", name, "output_path", path)

		var matches bool
		for _, secretName := range secretNames {
			if secretName == name {
				matches = true
				break
			}
		}

		if !matches {
			log.Info("Secret has been removed from config, deleting it from disk")
			if err := os.Remove(path); err != nil {
				return err
			}
		}
	}

	return nil
}
