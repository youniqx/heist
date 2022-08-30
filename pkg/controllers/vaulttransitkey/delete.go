package vaulttransitkey

import (
	"context"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) finalizeTransitKey(ctx context.Context, key *heistv1alpha1.VaultTransitKey) (ctrl.Result, error) {
	meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "transit key is being deleted",
	})

	engine, err := r.getEngineForKey(ctx, key)
	if err != nil {
		engine = &heistv1alpha1.VaultTransitEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Spec.Engine,
				Namespace: key.Namespace,
			},
			Spec: heistv1alpha1.VaultTransitEngineSpec{},
		}
	}

	if err := r.VaultAPI.DeleteTransitKey(engine, key); err != nil {
		r.Recorder.Eventf(key, "Warning", "ErrorDuringDeletion", "Failed to delete the key %s", key.Name)
		return common.Requeue, err
	}
	r.Recorder.Eventf(key, "Normal", "TransitKeyDeleted", "The key %s has been deleted", key.Name)

	if err := r.deletePolicy(key); err != nil {
		r.Recorder.Eventf(key, "Warning", "ErrorDuringDeletion", "Failed to delete policies for key %s", key.Name)
		return common.Requeue, err
	}
	r.Recorder.Eventf(key, "Normal", "TransitKeyPoliciesDeleted", "Policies for key %s have been deleted", key.Name)

	controllerutil.RemoveFinalizer(key, common.YouniqxFinalizer)

	return ctrl.Result{}, nil
}
