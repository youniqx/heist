package vaultkvsecret

import (
	"context"
	"errors"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateSecret(ctx context.Context, secret *heistv1alpha1.VaultKVSecret) (ctrl.Result, error) {
	engine, err := r.getEngineForSecret(ctx, secret)
	if err != nil {
		r.Recorder.Eventf(secret, "Warning", "EngineDoesNotExist", "Secret engine %s does not exist", secret.Spec.Engine)
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Referenced engine not found: %v", err),
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if meta.IsStatusConditionFalse(engine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(secret, "Normal", "EngineNotProvisionedYet", "Secret engine %s has not been provisioned yet. Waiting for it to settle...", engine.Name)
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  "waiting",
			Message: "Referenced engine is not provisioned yet",
		})
		return common.Requeue, fmt.Errorf("engine not provisioned yet")
	}

	r.attachFinalizer(secret)

	decryptError := ErrDecryptFailed.Copy()
	desired, current, err := r.determineState(engine, secret)
	switch {
	case errors.As(err, &decryptError):
		r.Recorder.Event(secret, "Warning", "DecryptFailed", decryptError.GetDetails())
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: decryptError.GetDetails(),
		})
		return common.Requeue, err
	case err != nil:
		r.Recorder.Eventf(secret, "Warning", "SecretStateDiffingFailed", "Failed to compare the current and desired state of secret %s", secret.Name)
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to diff current and desired secret state: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.performVaultReconciliation(desired, current); err != nil {
		r.Recorder.Eventf(secret, "Warning", "VaultReconciliationFailed", "Failed to roll out changes for secret %s to Vault", secret.Name)
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to apply changes in Vault: %v", err),
		})
		return common.Requeue, err
	}

	r.updateCurrentStateInSecret(secret, desired)

	return ctrl.Result{}, nil
}

func (r *Reconciler) updateCurrentStateInSecret(secret *heistv1alpha1.VaultKVSecret, newState *deployedSecret) {
	if meta.IsStatusConditionFalse(secret.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(secret, "Normal", "ProvisioningSuccessful", "Secret %s has been provisioned in engine %s", secret.Name, secret.Spec.Engine)
	}

	secret.Status.ReadOnlyPolicyName = newState.Policy.Name
	secret.Status.Engine = string(newState.Engine)
	secret.Status.Path = newState.Secret.Path
	secret.Status.Fields = newState.EncryptedFields

	meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "Secret has been provisioned",
	})
}

func (r *Reconciler) attachFinalizer(secret *heistv1alpha1.VaultKVSecret) {
	if !controllerutil.ContainsFinalizer(secret, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(secret, common.YouniqxFinalizer)
		r.Recorder.Event(secret, "Normal", "FinalizerAttached", "The finalizer has been attached to this object")
	}
}
