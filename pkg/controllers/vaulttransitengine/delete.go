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

func (r *Reconciler) finalizeEngine(_ context.Context, engine *heistv1alpha1.VaultTransitEngine) (ctrl.Result, error) {
	meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "Engine is being deleted",
	})

	if err := r.VaultAPI.DeleteEngine(engine); err != nil {
		r.Recorder.Eventf(engine, "Warning", "ErrorDuringDeletion", "Failed to delete the engine %s", engine.Name)
		meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to delete Transit Engine from Vault: %v", err),
		})
		return common.Requeue, err
	}

	r.Recorder.Eventf(engine, "Normal", "EngineDeleted", "The engine %s has been deleted", engine.Name)
	controllerutil.RemoveFinalizer(engine, common.YouniqxFinalizer)

	return ctrl.Result{}, nil
}
