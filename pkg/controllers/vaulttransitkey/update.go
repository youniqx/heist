package vaulttransitkey

import (
	"context"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateTransitKey(ctx context.Context, key *heistv1alpha1.VaultTransitKey) (ctrl.Result, error) {
	if key.Status.AppliedSpec.Type == "" {
		key.Status.AppliedSpec = key.Spec
	}

	if !controllerutil.ContainsFinalizer(key, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(key, common.YouniqxFinalizer)
		r.Recorder.Eventf(key, "Normal", "FinalizerAttached", "Attached finalizer to key %s", key.Name)
	}

	engine, err := r.getEngineForKey(ctx, key)
	if err != nil {
		r.Recorder.Eventf(key, "Warning", "EngineDoesNotExist", "Transit engine %s does not exist", key.Spec.Engine)
		meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Referenced TransitEngine not found: %v", err),
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if meta.IsStatusConditionFalse(engine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(key, "Normal", "EngineNotProvisionedYet", "Transit engine %s has not been provisioned yet. Waiting for it to settle...", engine.Name)
		meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  "waiting",
			Message: "Referenced engine is not provisioned yet",
		})
		return common.Requeue, fmt.Errorf("engine not provisioned yet")
	}

	if hasIncompatibleChanges(key) {
		oldKey := key.DeepCopy()
		oldKey.Spec = key.Status.AppliedSpec

		oldEngine, err := r.getEngineForKey(ctx, oldKey)
		if err != nil {
			r.Recorder.Eventf(key, "Warning", "OldEngineDoesNotExist", "Previously used Transit engine %s does not exist, can't cleanup", oldKey.Spec.Engine)
			return common.Requeue, fmt.Errorf("old engine doesn't exist, can't cleanup")
		}

		if err := r.VaultAPI.DeleteTransitKey(oldEngine, oldKey); err != nil {
			r.Recorder.Eventf(key, "Warning", "CantDeleteOldKey", "Previously applied transit key %s can't be deleted", oldKey.Name)
			meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
				Type:    heistv1alpha1.Conditions.Types.Provisioned,
				Status:  metav1.ConditionFalse,
				Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
				Message: "Failed to update transit key to newest spec",
			})
			return common.Requeue, fmt.Errorf("old key not deleted, can't cleanup")
		}
	}
	key.Status.AppliedSpec = key.Spec

	if err := r.VaultAPI.UpdateTransitKey(engine, key); err != nil {
		r.Recorder.Eventf(key, "Warning", "ProvisioningFailed", "Failed to provision key %s", key.Name)
		return common.Requeue, err
	}

	if err := r.updatePolicy(engine, key); err != nil {
		r.Recorder.Eventf(key, "Warning", "ProvisioningFailed", "Failed to provision policies for key %s", key.Name)
		return common.Requeue, err
	}

	if meta.IsStatusConditionFalse(key.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(key, "Normal", "ProvisioningSuccessful", "TransitKey %s has been provisioned", key.Name)
		meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionTrue,
			Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
			Message: "TransitKey has been provisioned",
		})
	}

	return ctrl.Result{}, nil
}

// hasIncompatibleChanges determines if the key spec has changed in a way
// that requires deleting the old key, and recreating a new one
// due to vault not being able to update to the new spec in-place.
func hasIncompatibleChanges(key *heistv1alpha1.VaultTransitKey) bool {
	if key.Status.AppliedSpec.Type == "" {
		return false
	}

	return hasChangedEngine(key) || hasChangedKeyType(key) || hasChangedAllowPlaintextBackup(key) || hasChangedExportable(key)
}

func hasChangedKeyType(key *heistv1alpha1.VaultTransitKey) bool {
	return key.Status.AppliedSpec.Type != key.Spec.Type
}

func hasChangedEngine(key *heistv1alpha1.VaultTransitKey) bool {
	return key.Status.AppliedSpec.Engine != "" && key.Status.AppliedSpec.Engine != key.Spec.Engine
}

func hasChangedAllowPlaintextBackup(key *heistv1alpha1.VaultTransitKey) bool {
	return key.Status.AppliedSpec.AllowPlaintextBackup != key.Spec.AllowPlaintextBackup
}

func hasChangedExportable(key *heistv1alpha1.VaultTransitKey) bool {
	return key.Status.AppliedSpec.Exportable != key.Spec.Exportable
}
