package agent

import (
	"reflect"

	"github.com/go-logr/logr"
	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault"
)

type config struct {
	API          vault.API
	ClientConfig *v1alpha1.VaultClientConfig
	Cache        *agentCache
}

func (c *config) AddToLogger(log logr.Logger) logr.Logger {
	return log.WithValues(
		"vault_address", c.ClientConfig.Spec.Address,
		"vault_role", c.ClientConfig.Spec.Role,
		"vault_auth_mount_path", c.ClientConfig.Spec.AuthMountPath,
		"kv_secret_count", len(c.ClientConfig.Spec.KvSecrets),
		"certificate_count", len(c.ClientConfig.Spec.Certificates),
		"ca_count", len(c.ClientConfig.Spec.CertificateAuthorities),
		"transit_key_count", len(c.ClientConfig.Spec.TransitKeys),
	)
}

func (c *config) SameVault(other *config) bool {
	if other == nil {
		return false
	}
	if c.ClientConfig.Spec.Address != other.ClientConfig.Spec.Address {
		return false
	}
	if c.ClientConfig.Spec.Role != other.ClientConfig.Spec.Role {
		return false
	}
	if !reflect.DeepEqual(c.ClientConfig.Spec.CACerts, other.ClientConfig.Spec.CACerts) {
		return false
	}
	if c.ClientConfig.Spec.AuthMountPath != other.ClientConfig.Spec.AuthMountPath {
		return false
	}
	return true
}
