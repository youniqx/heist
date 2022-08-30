package vaultkvsecret

import (
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) finalizeSecret(secret *heistv1alpha1.VaultKVSecret) (ctrl.Result, error) {
	meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "secret is being deleted",
	})

	if controllerutil.ContainsFinalizer(secret, common.YouniqxFinalizer) {
		if err := r.applyFinalizerSecret(secret); err != nil {
			return common.Requeue, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) applyFinalizerSecret(secret *heistv1alpha1.VaultKVSecret) error {
	meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "secret is being deleted",
	})

	currentState, err := r.determineCurrentState(secret)
	if err != nil {
		r.Recorder.Event(secret, "Warning", "CurrentStateResolutionFailed", "Failed to determine current state of secret")
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to determine current state of secret: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeletePolicy(currentState.Policy); err != nil {
		r.Recorder.Event(secret, "Warning", "ErrorDuringPolicyDeletion", "Failed to delete policy for secret from Vault.")
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete policy for secret from Vault: %v", err),
		})
		return err
	}

	if err := r.VaultAPI.DeleteKvSecret(currentState.Engine, currentState.Secret); err != nil {
		r.Recorder.Eventf(secret, "Warning", "ErrorDuringDeletion", "Failed to delete the secret %s from engine %s", secret.Name, secret.Spec.Engine)
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete the secret: %v", err),
		})

		return err
	}

	controllerutil.RemoveFinalizer(secret, common.YouniqxFinalizer)
	r.Recorder.Eventf(secret, "Normal", "FinalizerRemoved", "Secret %s is now ready for deletion from Kubernetes.", secret.Name)
	return nil
}
