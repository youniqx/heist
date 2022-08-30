package vaultbinding

import (
	"context"

	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/managed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateClientConfig(ctx context.Context, info *BindingInfo) error {
	kvSecrets, err := r.getKvSecretsForConfig(ctx, info)
	if err != nil {
		return err
	}

	caSecrets, err := r.getCertificateAuthoritySecretsForConfig(ctx, info)
	if err != nil {
		return err
	}

	certificates, err := r.getCertificateConfigFromBinding(ctx, info)
	if err != nil {
		return err
	}

	transitKeys, err := r.getTransitKeyConfigFromBinding(ctx, info)
	if err != nil {
		return err
	}

	config := &v1alpha1.VaultClientConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      info.Spec.Subject.Name,
			Namespace: info.Binding.Namespace,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, config, func() error {
		_ = controllerutil.SetControllerReference(info.Binding, config, r.Scheme)

		config.Spec = v1alpha1.VaultClientConfigSpec{
			Address:                r.VaultAPI.GetAddress(),
			Role:                   info.VaultRoleName,
			CACerts:                r.VaultAPI.GetCACerts(),
			AuthMountPath:          managed.KubernetesAuthPath,
			CertificateAuthorities: caSecrets,
			KvSecrets:              kvSecrets,
			Certificates:           certificates,
			TransitKeys:            transitKeys,
			Templates:              info.Spec.Agent,
		}

		return nil
	})
	if err != nil {
		return err
	}

	switch result {
	case controllerutil.OperationResultNone:
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(info.Binding, "Normal", "VaultClientConfig", "VaultClientConfig %s has been created", config.Name)
	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(info.Binding, "Normal", "VaultClientConfig", "VaultClientConfig %s has been updated", config.Name)
	case controllerutil.OperationResultUpdatedStatus:
	case controllerutil.OperationResultUpdatedStatusOnly:
	}

	return nil
}

func (r *Reconciler) getKvSecretsForConfig(ctx context.Context, info *BindingInfo) ([]*v1alpha1.VaultKVSecretRef, error) {
	secrets := make([]*v1alpha1.VaultKVSecretRef, 0, len(info.Spec.KVSecrets))
	for _, kv := range info.Spec.KVSecrets {
		secret := &v1alpha1.VaultKVSecret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kv.Name,
				Namespace: info.Binding.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(secret), secret); err != nil {
			return nil, err
		}

		engine := &v1alpha1.VaultKVSecretEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secret.Spec.Engine,
				Namespace: info.Binding.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(engine), engine); err != nil {
			return nil, err
		}

		enginePath, err := engine.GetMountPath()
		if err != nil {
			return nil, err
		}

		secretPath, err := secret.GetSecretPath()
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, &v1alpha1.VaultKVSecretRef{
			Name:         secret.Name,
			EnginePath:   enginePath,
			SecretPath:   secretPath,
			Capabilities: kv.Capabilities,
		})
	}
	return secrets, nil
}

func (r *Reconciler) getCertificateAuthoritySecretsForConfig(ctx context.Context, info *BindingInfo) ([]*v1alpha1.VaultCertificateAuthorityRef, error) {
	authorities := make([]*v1alpha1.VaultCertificateAuthorityRef, 0, len(info.Spec.CertificateAuthorities))
	for _, authority := range info.Spec.CertificateAuthorities {
		ca := &v1alpha1.VaultCertificateAuthority{
			ObjectMeta: metav1.ObjectMeta{
				Name:      authority.Name,
				Namespace: info.Binding.Namespace,
			},
		}
		if err := r.Get(ctx, client.ObjectKeyFromObject(ca), ca); err != nil {
			return nil, err
		}

		mountPath, err := ca.GetMountPath()
		if err != nil {
			return nil, err
		}

		authorities = append(authorities, &v1alpha1.VaultCertificateAuthorityRef{
			Name:         authority.Name,
			EnginePath:   mountPath,
			Capabilities: authority.Capabilities,
			KVSecrets: v1alpha1.VaultCertificateAuthorityKVSecretRef{
				EnginePath:        common.InternalKvEngineMountPath,
				PublicSecretPath:  common.GetCAInfoSecretPath(ca),
				PrivateSecretPath: common.GetCAPrivateKeySecretPath(ca),
			},
		})
	}
	return authorities, nil
}

func (r *Reconciler) getCertificateConfigFromBinding(ctx context.Context, info *BindingInfo) ([]*v1alpha1.VaultCertificateRef, error) {
	certificates := make([]*v1alpha1.VaultCertificateRef, 0, len(info.Spec.CertificateRoles))
	for _, certificate := range info.Spec.CertificateRoles {
		cert := &v1alpha1.VaultCertificateRole{
			ObjectMeta: metav1.ObjectMeta{
				Name:      certificate.Name,
				Namespace: info.Binding.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(cert), cert); err != nil {
			return nil, err
		}

		issuer := &v1alpha1.VaultCertificateAuthority{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cert.Spec.Issuer,
				Namespace: cert.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(issuer), issuer); err != nil {
			return nil, err
		}

		enginePath, err := issuer.GetMountPath()
		if err != nil {
			return nil, err
		}

		roleName, err := cert.GetRoleName()
		if err != nil {
			return nil, err
		}

		certificates = append(certificates, &v1alpha1.VaultCertificateRef{
			Name:         cert.Name,
			EnginePath:   enginePath,
			RoleName:     roleName,
			Capabilities: certificate.Capabilities,
		})
	}
	return certificates, nil
}

func (r *Reconciler) getTransitKeyConfigFromBinding(ctx context.Context, info *BindingInfo) ([]*v1alpha1.VaultTransitKeyRef, error) {
	transitKeys := make([]*v1alpha1.VaultTransitKeyRef, 0, len(info.Spec.CertificateRoles))
	for _, key := range info.Spec.TransitKeys {
		transitKey := &v1alpha1.VaultTransitKey{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: info.Binding.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(transitKey), transitKey); err != nil {
			return nil, err
		}

		transitEngine := &v1alpha1.VaultTransitEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      transitKey.Spec.Engine,
				Namespace: transitKey.Namespace,
			},
		}

		if err := r.Get(ctx, client.ObjectKeyFromObject(transitEngine), transitEngine); err != nil {
			return nil, err
		}

		enginePath, err := transitEngine.GetMountPath()
		if err != nil {
			return nil, err
		}

		keyName, err := transitKey.GetTransitKeyName()
		if err != nil {
			return nil, err
		}

		transitKeys = append(transitKeys, &v1alpha1.VaultTransitKeyRef{
			Name:         transitKey.Name,
			EnginePath:   enginePath,
			KeyName:      keyName,
			Capabilities: key.Capabilities,
		})
	}
	return transitKeys, nil
}
