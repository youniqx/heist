package vaultsyncsecret

import (
	"context"
	"errors"
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// Deprecated: Use syncFromAnnotation instead.
	deprecatedSyncFromAnnotation = "youniqx.com/sync-from"
	syncFromAnnotation           = "heist.youniqx.com/sync-from"
)

//nolint:cyclop
func (r *Reconciler) updateSync(ctx context.Context, sync *heistv1alpha1.VaultSyncSecret) (ctrl.Result, error) {
	r.attachFinalizer(sync)

	spec := sync.Spec

	var secretNamespace string
	switch {
	case spec.Target.Namespace == sync.Namespace, spec.Target.Namespace == "":
		secretNamespace = sync.Namespace
	case slices.Contains(r.NamespaceAllowList, spec.Target.Namespace):
		secretNamespace = spec.Target.Namespace
	default:
		r.Recorder.Eventf(sync, "Warning", "Misconfiguration", "Secret namespace %s is not allowed", spec.Target.Namespace)
		meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Namespace %s of secret is not allowed", spec.Target.Namespace),
		})
		return common.Requeue, nil
	}

	if hasMovedTargetSecret(sync) {
		if err := r.deleteOutdatedSecret(ctx, sync); err != nil {
			return common.Requeue, err
		}
	}
	sync.Status.AppliedSpec = spec

	renewalInterval, expectedData, err := r.FetchData(ctx, sync)
	if err != nil {
		meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Failed to fetch requested values: %v", err),
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	syncFromAnnotationValue := getSyncFromAnnotationValue(sync)

	target := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Target.Name,
			Namespace: secretNamespace,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, target, func() error {
		if target.Annotations == nil {
			target.Annotations = make(map[string]string)
		}

		if value, _ := common.GetAnnotationValue(target, deprecatedSyncFromAnnotation, syncFromAnnotation); value != syncFromAnnotationValue && target.UID != "" {
			return ErrAlreadyOwned
		}

		target.Type = v1.SecretType(pickFirstNonEmptyValue(string(spec.Target.Type), string(v1.SecretTypeOpaque)))
		delete(target.Annotations, deprecatedSyncFromAnnotation)
		target.Annotations[syncFromAnnotation] = syncFromAnnotationValue
		target.Data = expectedData

		if spec.Target.AdditionalAnnotations != nil {
			for key, value := range spec.Target.AdditionalAnnotations {
				target.Annotations[key] = value
			}
		}

		if spec.Target.AdditionalLabels != nil {
			if target.Labels == nil {
				target.Labels = make(map[string]string)
			}

			for key, value := range spec.Target.AdditionalLabels {
				target.Labels[key] = value
			}
		}

		_ = controllerutil.SetControllerReference(sync, target, r.Scheme)

		return nil
	})

	switch {
	case errors.Is(err, ErrAlreadyOwned):
		r.Recorder.Eventf(sync, "Warning", "Misconfiguration", "Secret %s/%s already exists and/or is managed by someone else", secretNamespace, spec.Target.Name)
		meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: "Secret already exists or is manged by someone else",
		})
		return common.Requeue, err
	case err != nil:
		r.Recorder.Eventf(sync, "Warning", "SyncSecretError", "Failed to create target secret %s/%s: %v", secretNamespace, spec.Target.Name, err)
		meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
			Message: fmt.Sprintf("Failed to create target secret: %v", err),
		})
		return common.Requeue, err
	}

	switch result {
	case controllerutil.OperationResultNone:
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(sync, "Normal", "Provisioned", "Secret %s/%s has been created", secretNamespace, spec.Target.Name)
	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(sync, "Normal", "Provisioned", "Secret %s/%s has been updated", secretNamespace, spec.Target.Name)
	case controllerutil.OperationResultUpdatedStatus:
	case controllerutil.OperationResultUpdatedStatusOnly:
	}

	meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "Secret has been synced",
	})

	return ctrl.Result{RequeueAfter: renewalInterval}, nil
}

func hasMovedTargetSecret(sync *heistv1alpha1.VaultSyncSecret) bool {
	spec := sync.Spec
	if sync.Status.AppliedSpec.Target.Namespace == "" {
		return false
	}
	if sync.Status.AppliedSpec.Target.Name == "" {
		return false
	}

	return sync.Status.AppliedSpec.Target.Namespace != spec.Target.Namespace || sync.Status.AppliedSpec.Target.Name != spec.Target.Name
}

func getSyncFromAnnotationValue(sync *heistv1alpha1.VaultSyncSecret) string {
	return fmt.Sprintf("%s/%s", sync.Namespace, sync.Name)
}

func (r *Reconciler) attachFinalizer(sync *heistv1alpha1.VaultSyncSecret) {
	if controllerutil.ContainsFinalizer(sync, common.YouniqxFinalizer) {
		return
	}
	controllerutil.AddFinalizer(sync, common.YouniqxFinalizer)
	r.Recorder.Eventf(sync, "Normal", "FinalizerAttached", "Attached a finalizer to VaultSyncSecret %s", sync.Name)
}

func (r *Reconciler) deleteOutdatedSecret(ctx context.Context, sync *heistv1alpha1.VaultSyncSecret) error {
	outdatedSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sync.Status.AppliedSpec.Target.Name,
			Namespace: sync.Status.AppliedSpec.Target.Namespace,
		},
	}

	switch err := r.Get(ctx, client.ObjectKeyFromObject(outdatedSecret), outdatedSecret); {
	case apierrors.IsNotFound(err):
		return nil
	case err != nil:
		return err
	default:
		if value, _ := common.GetAnnotationValue(outdatedSecret, deprecatedSyncFromAnnotation, syncFromAnnotation); value != getSyncFromAnnotationValue(sync) {
			return nil
		}
	}

	if err := r.Delete(ctx, outdatedSecret); err != nil {
		return client.IgnoreNotFound(err)
	}

	return nil
}
