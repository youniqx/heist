package vaultcertificateauthority

import (
	"context"
	"fmt"
	"path/filepath"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	"github.com/youniqx/heist/pkg/vault/pki"
	"github.com/youniqx/heist/pkg/vault/policy"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) performUpdate(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) (ctrl.Result, error) {
	if err := r.attachFinalizer(ctx, ca); err != nil {
		r.Recorder.Eventf(ca, "Warning", "AttachFinalizerFailed", "Failed to attach the finalizer to ca %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
			Message: fmt.Sprintf("Finalizer could not be attached to this VaultCertificateAuthority: %v", err),
		})
		r.Log.Error(err, "unable to attach finalizer")
		return common.Requeue, err
	}

	rootCA, err := common.FindRootCA(ctx, r.Client, ca)
	if err != nil {
		r.Recorder.Eventf(ca, "Warning", "FailedCAUpdate", "Failed to determine root ca for %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to find root ca: %v", err),
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	info, err := r.updateCAs(ctx, ca)
	if err != nil {
		r.Recorder.Eventf(ca, "Warning", "FailedCAUpdate", "Failed to update the ca %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Could not update CA to desired config: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.persistCAData(ca, rootCA, info); err != nil {
		r.Recorder.Eventf(ca, "Warning", "FailedCAUpdate", "Failed to persist ca data for %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to persist ca data: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.updatePolicy(ca); err != nil {
		r.Recorder.Eventf(ca, "Warning", "FailedCAUpdate", "Failed to update policies for %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to update policies: %v", err),
		})
		return common.Requeue, err
	}

	if meta.IsStatusConditionFalse(ca.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(ca, "Normal", "ProvisioningSuccessful", "CertificateAuthority %s has been provisioned", ca.Name)
	}

	meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "CertificateAuthority has been provisioned",
	})

	return ctrl.Result{}, nil
}

func (r *Reconciler) updateCAs(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) (*pki.CAInfo, error) {
	if ca.Spec.Issuer == "" {
		return r.updateRootCA(ca)
	}

	return r.updateIntermediateCA(ctx, ca)
}

func (r *Reconciler) updatePolicy(ca *heistv1alpha1.VaultCertificateAuthority) error {
	publicPolicy := &policy.Policy{
		Name: common.GetPolicyNameForCertificateAuthorityPublicInfo(ca),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(common.InternalKvEngineMountPath, "data", common.GetCAInfoSecretPath(ca)),
				Capabilities: []policy.Capability{
					policy.ReadCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(publicPolicy); err != nil {
		return err
	}

	privatePolicy := &policy.Policy{
		Name: common.GetPolicyNameForCertificateAuthorityPrivateInfo(ca),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(common.InternalKvEngineMountPath, "data", common.GetCAPrivateKeySecretPath(ca)),
				Capabilities: []policy.Capability{
					policy.ReadCapability,
				},
			},
		},
	}

	return r.VaultAPI.UpdatePolicy(privatePolicy)
}

func (r *Reconciler) updateRootCA(ca *heistv1alpha1.VaultCertificateAuthority) (*pki.CAInfo, error) {
	initialized, err := r.VaultAPI.IsPKIEngineInitialized(ca)
	if err != nil {
		return nil, err
	}

	entity, err := r.toVaultCAEntity(ca)
	if err != nil {
		return nil, err
	}

	if initialized {
		return nil, r.VaultAPI.UpdateRootCA(entity)
	}

	mode := pki.ModeInternal
	if ca.Spec.Settings.Exported {
		mode = pki.ModeExported
	}

	info, err := r.VaultAPI.CreateRootCA(mode, entity)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (r *Reconciler) updateIntermediateCA(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) (*pki.CAInfo, error) {
	initialized, err := r.VaultAPI.IsPKIEngineInitialized(ca)
	if err != nil {
		return nil, err
	}

	entity, err := r.toVaultCAEntity(ca)
	if err != nil {
		return nil, err
	}

	issuer := &heistv1alpha1.VaultCertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ca.Spec.Issuer,
			Namespace: ca.Namespace,
		},
	}

	if err := r.Get(ctx, client.ObjectKeyFromObject(issuer), issuer); err != nil {
		return nil, err
	}

	if initialized {
		return nil, r.VaultAPI.UpdateIntermediateCA(issuer, entity)
	}

	mode := pki.ModeInternal
	if ca.Spec.Settings.Exported {
		mode = pki.ModeExported
	}

	info, err := r.VaultAPI.CreateIntermediateCA(mode, issuer, entity)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (r *Reconciler) persistCAData(ca *heistv1alpha1.VaultCertificateAuthority, rootCA *heistv1alpha1.VaultCertificateAuthority, info *pki.CAInfo) error {
	if info == nil {
		return nil
	}

	if err := r.tryPersistingCAData(ca, rootCA, info); err != nil {
		r.Log.Info("Failed to persist CA data, rolling back changes to certificate authority")
		if err := r.VaultAPI.DeleteEngine(ca); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) tryPersistingCAData(ca *heistv1alpha1.VaultCertificateAuthority, rootCA *heistv1alpha1.VaultCertificateAuthority, info *pki.CAInfo) error {
	if err := r.VaultAPI.UpdateKvEngine(common.InternalKvEngine); err != nil {
		return err
	}

	var certificateChainPEM, fullCertificateChainPEM string

	if rootCA != ca {
		rootPEM, err := r.VaultAPI.ReadCACertificatePEM(rootCA)
		if err != nil {
			return err
		}
		certificateChainPEM = info.CertificateChain
		fullCertificateChainPEM = fmt.Sprintf("%s\n%s", info.Certificate, rootPEM)
	} else {
		certificateChainPEM = ""
		fullCertificateChainPEM = info.Certificate
	}

	publicSecret := &kvsecret.KvSecret{
		Path: common.GetCAInfoSecretPath(ca),
		Fields: map[string]string{
			common.CAIssuerField:               info.IssuingCertificateAuthority,
			common.CACertificateField:          info.Certificate,
			common.CACertificateChainField:     certificateChainPEM,
			common.CACertificateFullChainField: fullCertificateChainPEM,
			common.CASerialNumberField:         info.SerialNumber,
		},
	}

	if err := r.VaultAPI.UpdateKvSecret(common.InternalKvEngine, publicSecret); err != nil {
		return err
	}

	privateSecret := &kvsecret.KvSecret{
		Path: common.GetCAPrivateKeySecretPath(ca),
		Fields: map[string]string{
			common.CAPrivateKeyField:     info.PrivateKey,
			common.CAPrivateKeyTypeField: string(info.PrivateKeyType),
		},
	}

	return r.VaultAPI.UpdateKvSecret(common.InternalKvEngine, privateSecret)
}

func (r *Reconciler) attachFinalizer(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) error {
	if !controllerutil.ContainsFinalizer(ca, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(ca, common.YouniqxFinalizer)
		if err := r.Update(ctx, ca); err != nil {
			return err
		}
		r.Recorder.Event(ca, "Normal", "FinalizerAttached", "The finalizer has been attached to this object")
	}

	return nil
}
