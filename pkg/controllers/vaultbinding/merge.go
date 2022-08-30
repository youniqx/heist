package vaultbinding

import (
	"context"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BindingInfo struct {
	VaultRoleName string
	K8sRoleName   string
	Binding       *heistv1alpha1.VaultBinding
	Policies      []core.PolicyName
	Spec          *heistv1alpha1.VaultBindingSpec
}

func (r *Reconciler) buildBindingInfo(ctx context.Context, binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) (*BindingInfo, error) {
	policies, err := r.fetchPoliciesForBinding(ctx, binding, spec)
	if err != nil {
		return nil, err
	}

	return &BindingInfo{
		VaultRoleName: getVaultRoleName(binding, spec),
		K8sRoleName:   fmt.Sprintf("%s-client-config", spec.Subject.Name),
		Spec:          spec,
		Binding:       binding,
		Policies:      policies,
	}, nil
}

func getVaultRoleName(binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) string {
	return fmt.Sprintf("managed.k8s.%s.%s", binding.Namespace, spec.Subject.Name)
}

//nolint:cyclop
func (r *Reconciler) fetchPoliciesForBinding(ctx context.Context, binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) ([]core.PolicyName, error) {
	result := make([]core.PolicyName, 0, len(spec.KVSecrets)+len(spec.CertificateAuthorities)+len(spec.CertificateRoles)+len(spec.Capabilities))

	for _, capability := range spec.Capabilities {
		if capability == heistv1alpha1.VaultBindingHeistCapabilityEncrypt {
			result = append(result, managed.EncryptPolicy)
		}
	}

	for _, kv := range spec.KVSecrets {
		secret := &heistv1alpha1.VaultKVSecret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: binding.Namespace,
				Name:      kv.Name,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
			return nil, err
		}

		for _, capability := range kv.Capabilities {
			if capability == heistv1alpha1.VaultBindingKVCapabilityRead {
				result = append(result, core.PolicyName(common.GetPolicyNameForSecret(secret)))
			}
		}
	}

	for _, authority := range spec.CertificateAuthorities {
		ca := &heistv1alpha1.VaultCertificateAuthority{
			ObjectMeta: metav1.ObjectMeta{
				Name:      authority.Name,
				Namespace: binding.Namespace,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(ca), ca); err != nil {
			return nil, err
		}

		for _, capability := range authority.Capabilities {
			switch capability {
			case heistv1alpha1.VaultBindingCertificateAuthorityCapabilityReadPublic:
				result = append(result, core.PolicyName(common.GetPolicyNameForCertificateAuthorityPublicInfo(ca)))
			case heistv1alpha1.VaultBindingCertificateAuthorityCapabilityReadPrivate:
				result = append(result, core.PolicyName(common.GetPolicyNameForCertificateAuthorityPrivateInfo(ca)))
			}
		}
	}

	for _, info := range spec.CertificateRoles {
		cert := &heistv1alpha1.VaultCertificateRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:      info.Name,
				Namespace: binding.Namespace,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(cert), cert); err != nil {
			return nil, err
		}

		for _, capability := range info.Capabilities {
			switch capability {
			case heistv1alpha1.VaultBindingCertificateCapabilityIssue:
				result = append(result, core.PolicyName(common.GetPolicyNameForCertificateIssuing(cert)))
			case heistv1alpha1.VaultBindingCertificateCapabilitySignCSR:
				result = append(result, core.PolicyName(common.GetPolicyNameForCertificateSignCSR(cert)))
			case heistv1alpha1.VaultBindingCertificateCapabilitySignVerbatim:
				result = append(result, core.PolicyName(common.GetPolicyNameForCertificateSignVerbatim(cert)))
			}
		}
	}

	for _, info := range spec.TransitKeys {
		key := &heistv1alpha1.VaultTransitKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:      info.Name,
				Namespace: binding.Namespace,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(key), key); err != nil {
			return nil, err
		}

		for _, capability := range info.Capabilities {
			switch capability {
			case heistv1alpha1.VaultBindingTransitKeyCapabilityEncrypt:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyEncrypt(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityDecrypt:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyDecrypt(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityRewrap:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyRewrap(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityDatakey:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyDatakey(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityHmac:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyHmac(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilitySign:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeySign(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityVerify:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyVerify(key)))
			case heistv1alpha1.VaultBindingTransitKeyCapabilityRead:
				result = append(result, core.PolicyName(common.GetPolicyNameForTransitKeyRead(key)))
			}
		}
	}

	return result, nil
}
