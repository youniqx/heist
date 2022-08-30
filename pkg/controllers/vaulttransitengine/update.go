package vaulttransitengine

import (
	"context"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateEngine(_ context.Context, engine *heistv1alpha1.VaultTransitEngine) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(engine, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(engine, common.YouniqxFinalizer)
		r.Recorder.Eventf(engine, "Normal", "FinalizerAttached", "Attached finalizer to engine %s", engine.Name)
	}

	if err := r.VaultAPI.UpdateTransitEngine(engine); err != nil {
		r.Recorder.Eventf(engine, "Warning", "ProvisioningFailed", "Failed to provision engine %s", engine.Name)
		meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionTrue,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to update transit engine: %v", err),
		})
		return common.Requeue, err
	}

	if meta.IsStatusConditionFalse(engine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(engine, "Normal", "ProvisioningSuccessful", "Engine %s has been provisioned", engine.Name)
		meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionTrue,
			Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
			Message: "Engine has been provisioned",
		})
	}

	return ctrl.Result{}, nil
}
