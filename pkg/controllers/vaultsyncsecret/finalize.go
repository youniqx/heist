package vaultsyncsecret

import (
	"context"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//nolint:cyclop
func (r *Reconciler) finalizeSync(ctx context.Context, sync *heistv1alpha1.VaultSyncSecret) (ctrl.Result, error) {
	var secretNamespace string
	deleteSecret := true
	spec := sync.Spec
	switch {
	case spec.Target.Namespace == sync.Namespace, spec.Target.Namespace == "":
		secretNamespace = sync.Namespace
	case slices.Contains(r.NamespaceAllowList, spec.Target.Namespace):
		secretNamespace = spec.Target.Namespace
	default:
		deleteSecret = false
	}

	target := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        spec.Target.Name,
			Namespace:   secretNamespace,
			Annotations: make(map[string]string),
		},
	}

	if deleteSecret {
		syncFromValue := getSyncFromAnnotationValue(sync)

		switch err := r.Get(ctx, client.ObjectKeyFromObject(target), target); {
		case apierrors.IsNotFound(err):
			deleteSecret = false
		case err == nil:
			if value, _ := common.GetAnnotationValue(target, deprecatedSyncFromAnnotation, syncFromAnnotation); value != syncFromValue {
				deleteSecret = false
			}
		default:
			return common.Requeue, err
		}
	}

	if deleteSecret {
		switch err := r.Delete(ctx, target); {
		case apierrors.IsNotFound(err):
		case err == nil:
			r.Recorder.Eventf(sync, "Normal", "Secret Deleted", "Secret %s/%s has been deleted", secretNamespace, sync.Spec.Target.Name)
		default:
			return common.Requeue, err
		}
	}

	r.detachFinalizer(sync)

	return ctrl.Result{}, nil
}

func (r *Reconciler) detachFinalizer(sync *heistv1alpha1.VaultSyncSecret) {
	if !controllerutil.ContainsFinalizer(sync, common.YouniqxFinalizer) {
		return
	}
	controllerutil.RemoveFinalizer(sync, common.YouniqxFinalizer)
	r.Recorder.Eventf(sync, "Normal", "FinalizerRemoved", "Finalizer has been removed from VaultSyncSecret %s", sync.Name)
}
