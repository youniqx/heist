package vaultcertificateauthority

import (
	"context"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/core"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) finalizeCA(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) (ctrl.Result, error) {
	r.Recorder.Event(ca, "Normal", "DeletionInProgress", "The VaultCertificateAuthority has been marked for deletion. CA will be deleted from Vault.")
	meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "VaultCertificateAuthority is in the process of being deleted from Vault",
	})

	if controllerutil.ContainsFinalizer(ca, common.YouniqxFinalizer) {
		if err := r.deleteCA(ctx, ca); err != nil {
			return common.Requeue, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteCA(ctx context.Context, ca *heistv1alpha1.VaultCertificateAuthority) error {
	if err := r.VaultAPI.DeleteEngine(ca); err != nil {
		r.Recorder.Event(ca, "Warning", "ErrorDuringCADeletion", "Failed to delete VaultCertificateAuthority from Vault.")
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete VaultCertificateAuthority from Vault: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeleteKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(ca))); err != nil {
		r.Recorder.Event(ca, "Warning", "ErrorDuringCADeletion", "Failed to delete private key from Vault.")
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete private key from Vault: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeleteKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAPrivateKeySecretPath(ca))); err != nil {
		r.Recorder.Event(ca, "Warning", "ErrorDuringCADeletion", "Failed to delete private key from Vault.")
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete private key from Vault: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForCertificateAuthorityPublicInfo(ca))); err != nil {
		r.Recorder.Event(ca, "Warning", "ErrorDuringCADeletion", "Failed to delete ca policy from Vault.")
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete ca policy from Vault: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForCertificateAuthorityPrivateInfo(ca))); err != nil {
		r.Recorder.Event(ca, "Warning", "ErrorDuringCADeletion", "Failed to delete ca policy from Vault.")
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete ca policy from Vault: %v", err),
		})
		return err
	}

	controllerutil.RemoveFinalizer(ca, common.YouniqxFinalizer)
	if err := r.Update(ctx, ca); err != nil {
		r.Recorder.Eventf(ca, "Warning", "RemoveFinalizerFailed", "Failed to remove finalizer from ca %s", ca.Name)
		meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
			Message: fmt.Sprintf("Finalizer could not be removed: %v", err),
		})
		return client.IgnoreNotFound(err)
	}

	r.Recorder.Eventf(ca, "Normal", "FinalizerRemoved", "VaultCertificateAuthority %s is now ready for deletion from Kubernetes.", ca.Name)
	meta.SetStatusCondition(&ca.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "VaultCertificateAuthority was successfully deleted from Vault",
	})
	return nil
}
