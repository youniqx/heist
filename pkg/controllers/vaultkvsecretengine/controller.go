/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vaultkvsecretengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-test/deep"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var ErrEngineExistsAfterDelete = errors.New("error exists after deletion")

// Reconciler reconciles a VaultKVSecretEngine object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	VaultAPI    vault.API
	Recorder    record.EventRecorder
	EventFilter predicate.Predicate
}

// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecretengines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecretengines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecretengines/finalizers,verbs=update

// Reconcile sets up the controller with the Manager.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaultkvsecretengine", req.NamespacedName)
	log.Info("reconciling for secret engine")

	engine := &heistv1alpha1.VaultKVSecretEngine{}
	if err := r.Get(ctx, req.NamespacedName, engine); err != nil { // nestif
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to fetch VaultKVSecretEngine")
		} else {
			err = r.checkAndCleanupStaleVaultKVEngine(req)
			if err != nil {
				return common.Requeue, err
			}
		}

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previous := engine.DeepCopy()

	setDefaultConditions(engine)

	result, err := r.performReconciliation(ctx, engine)

	if deep.Equal(previous.Status, engine.Status) != nil {
		if err := r.Status().Update(ctx, engine); err != nil {
			return common.Requeue, err
		}
	}

	if deep.Equal(previous.Finalizers, engine.Finalizers) != nil {
		if err := r.Update(ctx, engine); err != nil {
			return common.Requeue, err
		}
	}

	return result, err
}

func (r *Reconciler) checkAndCleanupStaleVaultKVEngine(req ctrl.Request) (err error) {
	log := r.Log.WithValues("vaultkvsecretengine", req.NamespacedName, "method", "checkAndCleanupStaleVaultKVEngine")

	log.Info("k8s vault engine object already is deleted, verifying no stale engine in vault is left")

	engine := &heistv1alpha1.VaultKVSecretEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	engineStillExists, err := r.VaultAPI.HasEngine(engine)
	if err != nil {
		log.Info("couldn't check if engine exists for deleted k8s manifest", "error", err)
		return err
	}

	if !engineStillExists {
		log.Info("engine was successfully deleted from k8s and vault, nothing todo")
		return nil
	}

	err = r.VaultAPI.DeleteEngine(engine)
	if err != nil {
		log.Info("couldn't delete stale vault engine", "error", err)
		return err
	}

	return nil
}

func setDefaultConditions(engine *heistv1alpha1.VaultKVSecretEngine) {
	if meta.FindStatusCondition(engine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) performReconciliation(ctx context.Context, engine *heistv1alpha1.VaultKVSecretEngine) (ctrl.Result, error) {
	if engine.GetDeletionTimestamp() != nil {
		return r.finalizeEngine(ctx, engine)
	}

	return r.updateEngine(engine)
}

func (r *Reconciler) updateEngine(engine *heistv1alpha1.VaultKVSecretEngine) (ctrl.Result, error) {
	r.attachFinalizer(engine)

	if err := r.VaultAPI.UpdateKvEngine(engine); err != nil {
		r.Recorder.Eventf(engine, "Warning", "ProvisioningFailed", "Failed to provision engine %s", engine.Name)
		return common.Requeue, err
	}

	meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "Engine has been provisioned",
	})

	return ctrl.Result{}, nil
}

func (r *Reconciler) attachFinalizer(engine *heistv1alpha1.VaultKVSecretEngine) {
	if !controllerutil.ContainsFinalizer(engine, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(engine, common.YouniqxFinalizer)
		r.Recorder.Eventf(engine, "Normal", "FinalizerAttached", "Attached finalizer to engine %s", engine.Name)
	}
}

func (r *Reconciler) finalizeEngine(ctx context.Context, engine *heistv1alpha1.VaultKVSecretEngine) (ctrl.Result, error) {
	meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "Engine is being deleted",
	})

	if controllerutil.ContainsFinalizer(engine, common.YouniqxFinalizer) {
		err := r.applyFinalizerEngine(ctx, engine)
		if err != nil {
			return common.Requeue, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) applyFinalizerEngine(ctx context.Context, engine *heistv1alpha1.VaultKVSecretEngine) error {
	log := r.Log.WithValues("method", "applyFinalizerEngine")

	path, err := engine.GetMountPath()
	if err != nil {
		return fmt.Errorf("couldn't get mount path for engine %s: %w", engine.Name, err)
	}

	log = log.WithValues("path", path)

	log.Info("finalizing secret engine")

	secrets := &heistv1alpha1.VaultKVSecretList{}
	if err := r.List(ctx, secrets, client.InNamespace(engine.Namespace)); err != nil {
		return err
	}

	log.Info("deleting engine in vault")

	if err := r.VaultAPI.DeleteEngine(engine); err != nil {
		log.Info("failed to delete engine", "error", err)
		r.Recorder.Eventf(engine, "Warning", "ErrorDuringDeletion", "Failed to delete the engine %s", engine.Name)
		// TODO: add condition
		return err
	}

	log.Info("verifying if vault api actually deleted the engine")

	engineStillExists, err := r.VaultAPI.HasEngine(engine)
	if err != nil {
		log.Info("failed verify deletion of engine", "error", err)
		r.Recorder.Eventf(engine, "Warning", "ErrorDuringDeletion", "Failed verify deletion of engine %s", engine.Name)
		// TODO: add condition
		return err
	}

	if engineStillExists {
		log.Info("engine still exists after it should have been deleted returning", "error", ErrEngineExistsAfterDelete)
		r.Recorder.Eventf(engine, "Warning", "ErrorDuringDeletion", "Engine %s still exists after deletion", engine.Name)
		return ErrEngineExistsAfterDelete
	}

	r.Recorder.Eventf(engine, "Normal", "EngineDeleted", "The engine %s has been deleted", engine.Name)
	controllerutil.RemoveFinalizer(engine, common.YouniqxFinalizer)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultKVSecretEngine{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}
