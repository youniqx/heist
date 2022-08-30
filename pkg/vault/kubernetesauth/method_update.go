package kubernetesauth

import (
	"errors"
	"net/http"
	"path/filepath"
	"reflect"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
)

type Config struct {
	KubernetesHost       string   `json:"kubernetes_host,omitempty"`
	Issuer               string   `json:"issuer,omitempty"`
	PemKeys              []string `json:"pem_keys,omitempty"`
	KubernetesCACert     string   `json:"kubernetes_ca_cert,omitempty"`
	TokenReviewerJWT     string   `json:"token_reviewer_jwt,omitempty"`
	DisableISSValidation bool     `json:"disable_iss_validation,omitempty"`
	DisableLocalCAJWT    bool     `json:"disable_local_ca_jwt,omitempty"`
}

//nolint:cyclop
func (k *kubernetesAuthAPI) UpdateKubernetesAuthMethod(method MethodEntity) error {
	log := k.Core.Log().WithValues("method", "UpdateKubernetesAuthMethod")

	path, err := method.GetMountPath()
	if err != nil {
		log.Info("failed to get mount path", "error", err)
		return core.ErrAPIError.WithDetails("failed to get auth mount path").WithCause(err)
	}

	log = log.WithValues("path", path)

	exists, err := k.Auth.HasAuthMethod(method)
	if err != nil {
		log.Info("failed to check if k8s auth method exists", "error", err)
		return core.ErrAPIError.WithDetails("failed to check if k8s auth method exists").WithCause(err)
	}

	if !exists {
		authMethod := &auth.Method{
			Path: path,
			Type: auth.MethodKubernetes,
		}

		if err := k.Auth.CreateAuthMethod(authMethod); err != nil {
			log.Info("failed to create k8s auth method", "error", err)
			return core.ErrAPIError.WithDetails("failed to create k8s auth method").WithCause(err)
		}
	}

	desiredConfig, err := method.GetMethodConfig()
	if err != nil {
		log.Info("failed to get k8s auth method config", "error", err)
		return core.ErrAPIError.WithDetails("failed to get k8s auth method config").WithCause(err)
	}

	configPath := filepath.Join("/v1/auth", path, "config")

	actualConfig := &Config{}
	if err := k.Core.MakeRequest(core.MethodGet, configPath, nil, httpclient.JSON(actualConfig)); err != nil {
		var responseError *core.VaultHTTPError
		if !errors.As(err, &responseError) || responseError.StatusCode != http.StatusNotFound {
			log.Info("unable to create auth method", "error", err)
			return core.ErrAPIError.WithDetails("failed to create auth method").WithCause(err)
		}
	}

	if reflect.DeepEqual(actualConfig, desiredConfig) {
		return nil
	}

	if err := k.Core.MakeRequest(core.MethodPost, configPath, httpclient.JSON(desiredConfig), nil); err != nil {
		log.Info("unable to create auth method", "error", err)
		return core.ErrAPIError.WithDetails("failed to create auth method").WithCause(err)
	}

	return nil
}
