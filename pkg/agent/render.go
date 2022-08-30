package agent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/pki"
)

type secretRenderer struct {
	ClientConfig *v1alpha1.VaultClientConfig
	Cache        *agentCache
}

func (r *secretRenderer) kvSecret(name string, field string) (string, error) {
	enginePath, secretPath, err := r.getKvSecretPaths(name)
	if err != nil {
		return "", err
	}

	secret, err := r.Cache.ReadKvSecret(enginePath, secretPath)
	if err != nil {
		return "", err
	}
	return secret.Fields[field], nil
}

func (r *secretRenderer) getKvSecretPaths(name string) (enginePath core.MountPathEntity, secretPath core.SecretPathEntity, err error) {
	for _, kvSecret := range r.ClientConfig.Spec.KvSecrets {
		if kvSecret.Name == name {
			return core.MountPath(kvSecret.EnginePath), core.SecretPath(kvSecret.SecretPath), nil
		}
	}

	return nil, nil, ErrNotFound.WithDetails(fmt.Sprintf("failed to find kv secret with name %s", name))
}

func (r *secretRenderer) caField(name string, field v1alpha1.VaultCertificateFieldType) (string, error) {
	info, err := r.getCAInfo(name)
	if err != nil {
		return "", err
	}

	switch field {
	case v1alpha1.VaultBindingCertificateFieldTypeFullCertChain:
		secret, err := r.Cache.ReadKvSecret(core.MountPath(info.KVSecrets.EnginePath), core.SecretPath(info.KVSecrets.PublicSecretPath))
		if err != nil {
			return "", err
		}
		return secret.Fields[common.CACertificateFullChainField], nil
	case v1alpha1.VaultBindingCertificateFieldTypeCertChain:
		secret, err := r.Cache.ReadKvSecret(core.MountPath(info.KVSecrets.EnginePath), core.SecretPath(info.KVSecrets.PublicSecretPath))
		if err != nil {
			return "", err
		}
		return secret.Fields[common.CACertificateChainField], nil
	case v1alpha1.VaultBindingCertificateFieldTypeCertificate:
		secret, err := r.Cache.ReadKvSecret(core.MountPath(info.KVSecrets.EnginePath), core.SecretPath(info.KVSecrets.PublicSecretPath))
		if err != nil {
			return "", err
		}
		return secret.Fields[common.CACertificateField], nil
	case v1alpha1.VaultBindingCertificateFieldTypePrivateKey:
		secret, err := r.Cache.ReadKvSecret(core.MountPath(info.KVSecrets.EnginePath), core.SecretPath(info.KVSecrets.PrivateSecretPath))
		if err != nil {
			return "", err
		}
		return secret.Fields[common.CAPrivateKeyField], nil
	}

	return "", nil
}

func (r *secretRenderer) getCAInfo(name string) (*v1alpha1.VaultCertificateAuthorityRef, error) {
	for _, authority := range r.ClientConfig.Spec.CertificateAuthorities {
		if authority.Name == name {
			return authority, nil
		}
	}

	return nil, ErrNotFound.WithDetails(fmt.Sprintf("failed to find configuration for CA %s", name))
}

func (r *secretRenderer) certField(name string, field v1alpha1.VaultCertificateFieldType) (string, error) {
	enginePath, roleName, issueOptions, err := r.getCertificatePaths(name)
	if err != nil {
		return "", err
	}

	certificate, err := r.Cache.IssueCertificate(enginePath, roleName, issueOptions)
	if err != nil {
		return "", err
	}

	switch field {
	case v1alpha1.VaultBindingCertificateFieldTypeFullCertChain:
		chain := make([]string, 0, len(certificate.CAChain)+1)
		chain = append(chain, certificate.Certificate)
		chain = append(chain, certificate.CAChain...)
		// TODO: Append root cert
		return strings.Join(chain, "\n"), nil
	case v1alpha1.VaultBindingCertificateFieldTypeCertChain:
		chain := make([]string, 0, len(certificate.CAChain)+1)
		chain = append(chain, certificate.Certificate)
		chain = append(chain, certificate.CAChain...)
		return strings.Join(chain, "\n"), nil
	case v1alpha1.VaultBindingCertificateFieldTypeCertificate:
		return certificate.Certificate, nil
	case v1alpha1.VaultBindingCertificateFieldTypePrivateKey:
		return certificate.PrivateKey, nil
	}

	return "", nil
}

func (r *secretRenderer) getCertificatePaths(name string) (enginePath core.MountPathEntity, role core.RoleNameEntity, options *pki.IssueCertOptions, err error) {
	certTemplate, err := r.findCertTemplate(name)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, certificate := range r.ClientConfig.Spec.Certificates {
		if certificate.Name == certTemplate.CertificateRole {
			return core.MountPath(certificate.EnginePath), core.RoleName(certificate.RoleName), &pki.IssueCertOptions{
				CommonName:        certTemplate.CommonName,
				DNSSans:           certTemplate.DNSSans,
				OtherSans:         certTemplate.OtherSans,
				IPSans:            certTemplate.IPSans,
				URISans:           certTemplate.URISans,
				TTL:               certTemplate.TTL.Duration,
				ExcludeCNFromSans: certTemplate.ExcludeCNFromSans,
			}, nil
		}
	}

	return nil, nil, nil, ErrNotFound.WithDetails(fmt.Sprintf("failed to find certificate referenced by certificate template %s", name))
}

func (r *secretRenderer) findCertTemplate(name string) (*v1alpha1.VaultCertificateTemplate, error) {
	for _, certificate := range r.ClientConfig.Spec.Templates.CertificateTemplates {
		if certificate.Alias != "" {
			if certificate.Alias == name {
				return &certificate, nil
			}
		} else if certificate.CertificateRole == name {
			return &certificate, nil
		}
	}
	return nil, ErrNotFound.WithDetails(fmt.Sprintf("failed to find certificate template with name %s", name))
}

func (r *secretRenderer) Render(secret v1alpha1.VaultBindingValueTemplate) (string, error) {
	tpl, err := template.New("secret").
		Funcs(map[string]interface{}{
			"kvSecret":  r.kvSecret,
			"certField": r.certField,
			"caField":   r.caField,
		}).
		Funcs(sprig.GenericFuncMap()).
		Parse(secret.Template)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, nil); err != nil {
		return "", err
	}

	return buffer.String(), nil
}
