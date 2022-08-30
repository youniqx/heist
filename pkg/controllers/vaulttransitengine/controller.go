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

package vaulttransitengine

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/go-test/deep"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconciler reconciles a VaultTransitEngine object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	VaultAPI    vault.API
	Recorder    record.EventRecorder
	EventFilter predicate.Predicate
}

// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitengines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitengines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitengines/finalizers,verbs=update

// Reconcile sets up the controller with the Manager.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaulttransitengine", req.NamespacedName)
	log.Info("reconciling for secret engine")

	engine := &heistv1alpha1.VaultTransitEngine{}
	if err := r.Get(ctx, req.NamespacedName, engine); err != nil {
		if err2 := client.IgnoreNotFound(err); err2 != nil { // nestif
			log.Error(err, "unable to fetch VaultTransitEngine")
		} else {
			errchkeng := r.checkAndCleanupStaleVaultTransitEngine(req)
			if errchkeng != nil {
				return common.Requeue, errchkeng
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

func (r *Reconciler) performReconciliation(ctx context.Context, engine *heistv1alpha1.VaultTransitEngine) (ctrl.Result, error) {
	if engine.GetDeletionTimestamp() != nil {
		return r.finalizeEngine(ctx, engine)
	}

	return r.updateEngine(ctx, engine)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultTransitEngine{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}

func setDefaultConditions(engine *heistv1alpha1.VaultTransitEngine) {
	if meta.FindStatusCondition(engine.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&engine.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) checkAndCleanupStaleVaultTransitEngine(req ctrl.Request) (err error) {
	log := r.Log.WithValues("vaulttransitengine", req.NamespacedName, "method", "checkAndCleanupStaleVaultTransitEngine")

	log.Info("k8s vault engine object already is deleted, verifying no stale engine in vault is left")

	engine := &heistv1alpha1.VaultTransitEngine{
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
