package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/erx"
)

var (
	// ErrNotYetSynced is returned when the config hasn't been synced yet.
	ErrNotYetSynced = erx.New("Heist Agent", "Config is not synced yet").
			WithDetails("config has not been synced to the agent yet, this may take up to a couple of seconds")
	// ErrAPIRequestFailed is returned when vault returned an unexpected response.
	ErrAPIRequestFailed = erx.New("Heist Agent", "REST requests to Vault could not be completed")
	// ErrNotFound is returned when a secret doesn't exist.
	ErrNotFound = erx.New("Heist Agent", "not found")
)

const (
	fileModeBase    = 8
	fileModeBitSize = 32
	fileModeDefault = 0o640
)

type Secret struct {
	Value      string
	Name       string
	OutputPath string
	Mode       os.FileMode
}

func (a *agent) GetClientSecret() *Secret {
	result := &Secret{
		Value:      "",
		Name:       "heist.json",
		OutputPath: filepath.Join(a.BasePath, "config.json"),
		Mode:       fileModeDefault,
	}

	conf := a.getConfig()
	if conf == nil {
		return result
	}

	spec := conf.ClientConfig.Spec.DeepCopy()
	spec.Templates = v1alpha1.VaultBindingAgentConfig{}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetIndent("", "\t")
	if err := encoder.Encode(spec); err != nil {
		return result
	}

	result.Value = buffer.String()

	return result
}

func (a *agent) ListSecrets() (names []string, err error) {
	log := a.Log.WithValues("operation", "ListSecrets")
	conf := a.getConfig()
	if conf == nil {
		log.Info("failed to list secrets because config is not synced yet")
		return nil, ErrNotYetSynced
	}

	log = conf.AddToLogger(log)

	result := make([]string, 0, len(conf.ClientConfig.Spec.Templates.Templates))
	for _, template := range conf.ClientConfig.Spec.Templates.Templates {
		result = append(result, template.Path)
	}

	log.Info("successfully listed secrets", "count", len(result))

	return result, nil
}

func (a *agent) FetchSecret(name string) (value *Secret, err error) {
	log := a.Log.WithValues("operation", "FetchSecret", "name", name)

	conf := a.getConfig()
	if conf == nil {
		log.Info("failed to fetch secret because config is not synced yet")
		return nil, ErrNotYetSynced
	}

	for _, secret := range conf.ClientConfig.Spec.Templates.Templates {
		if name == secret.Path {
			return a.renderTemplate(secret)
		}
	}

	return nil, ErrNotFound.WithDetails(fmt.Sprintf("secret with name %s not found", name))
}

func (a *agent) renderTemplate(secret v1alpha1.VaultBindingValueTemplate) (*Secret, error) {
	log := a.Log.WithValues("operation", "renderTemplate")
	conf := a.getConfig()
	if conf == nil {
		log.Info("failed to render template because config is not synced yet")
		return nil, ErrNotYetSynced
	}

	renderer := &secretRenderer{
		ClientConfig: conf.ClientConfig,
		Cache:        conf.Cache,
	}

	value, err := renderer.Render(secret)
	if err != nil {
		return nil, err
	}

	return &Secret{
		Value:      value,
		Name:       secret.Path,
		OutputPath: a.createOutputPath(secret.Path),
		Mode:       parseMode(secret.Mode),
	}, nil
}

func (a *agent) createOutputPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(a.BasePath, "secrets", path)
}

func parseMode(mode string) os.FileMode {
	if mode == "" {
		return fileModeDefault
	}

	result, err := strconv.ParseInt(mode, fileModeBase, fileModeBitSize)
	if err != nil {
		return fileModeDefault
	}

	return os.FileMode(result)
}
