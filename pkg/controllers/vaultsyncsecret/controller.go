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

package vaultsyncsecret

import (
	"context"

	"github.com/go-test/deep"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles a VaultSyncSecret object.
type Reconciler struct {
	client.Client
	Scheme             *runtime.Scheme
	VaultAPI           vault.API
	Recorder           record.EventRecorder
	EventFilter        operator.AnnotationFilter
	NamespaceAllowList []string
}

//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultsyncsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultsyncsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultsyncsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultSyncSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	sync := &heistv1alpha1.VaultSyncSecret{}
	if err := r.Get(ctx, req.NamespacedName, sync); err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			logger.Info("unable to fetch VaultSyncSecret", "error", err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previous := sync.DeepCopy()

	setDefaultConditions(sync)

	result, err := r.performReconciliation(ctx, sync)

	if deep.Equal(previous.Status, sync.Status) != nil {
		if err := r.Status().Update(ctx, sync); err != nil {
			return common.Requeue, err
		}
	}

	if deep.Equal(previous.Finalizers, sync.Finalizers) != nil {
		if err := r.Update(ctx, sync); err != nil {
			return common.Requeue, err
		}
	}

	return result, err
}

func setDefaultConditions(sync *heistv1alpha1.VaultSyncSecret) {
	if meta.FindStatusCondition(sync.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&sync.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func pickFirstNonEmptyValue(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (r *Reconciler) performReconciliation(ctx context.Context, sync *heistv1alpha1.VaultSyncSecret) (ctrl.Result, error) {
	if sync.GetDeletionTimestamp() != nil {
		return r.finalizeSync(ctx, sync)
	}

	return r.updateSync(ctx, sync)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultSyncSecret{}).
		Watches(&source.Kind{Type: &v1.Secret{}}, &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &heistv1alpha1.VaultSyncSecret{}}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}
