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

package vaultkvsecret

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-test/deep"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler reconciles a VaultKVSecret object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	VaultAPI    vault.API
	Recorder    record.EventRecorder
	EventFilter predicate.Predicate
}

// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecrets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultkvsecrets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultKVSecret object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaultkvsecret", req.NamespacedName)
	log.Info("reconciling for vault secret")

	secret := &heistv1alpha1.VaultKVSecret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		if err2 := client.IgnoreNotFound(err); err2 != nil {
			log.Error(err, "unable to fetch VaultKVSecret")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previous := secret.DeepCopy()

	setDefaultConditions(secret)

	result, err := r.performReconciliation(ctx, secret)

	if deep.Equal(previous.Status, secret.Status) != nil {
		if err := r.Status().Update(ctx, secret); err != nil {
			return common.Requeue, err
		}
	}

	if deep.Equal(previous.Finalizers, secret.Finalizers) != nil {
		if err := r.Update(ctx, secret); err != nil {
			return common.Requeue, err
		}
	}

	return result, err
}

func setDefaultConditions(secret *heistv1alpha1.VaultKVSecret) {
	if meta.FindStatusCondition(secret.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&secret.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) performReconciliation(ctx context.Context, secret *heistv1alpha1.VaultKVSecret) (ctrl.Result, error) {
	if secret.GetDeletionTimestamp() != nil {
		return r.finalizeSecret(secret)
	}

	return r.updateSecret(ctx, secret)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultKVSecret{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Watches(
			&heistv1alpha1.VaultKVSecretEngine{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				engine, ok := object.(*heistv1alpha1.VaultKVSecretEngine)
				if !ok {
					return nil
				}

				secrets := heistv1alpha1.VaultKVSecretList{}
				for i := 0; i < 3; i++ {
					if err := mgr.GetClient().List(context.TODO(), &secrets, &client.ListOptions{Namespace: engine.Namespace}); err != nil {
						time.Sleep(time.Second)
						continue
					}
					requests := make([]reconcile.Request, 0, len(secrets.Items))
					for _, secret := range secrets.Items {
						if secret.Spec.Engine != engine.Name {
							continue
						}

						requests = append(requests, reconcile.Request{
							NamespacedName: types.NamespacedName{
								Name:      secret.Name,
								Namespace: secret.Namespace,
							},
						})
					}
					return requests
				}
				return nil
			}),
		).
		Complete(r)
}

func (r *Reconciler) getEngineForSecret(ctx context.Context, secret *heistv1alpha1.VaultKVSecret) (*heistv1alpha1.VaultKVSecretEngine, error) {
	engine := &heistv1alpha1.VaultKVSecretEngine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secret.Namespace,
			Name:      secret.Spec.Engine,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(engine), engine); err != nil {
		return nil, err
	}
	return engine, nil
}
