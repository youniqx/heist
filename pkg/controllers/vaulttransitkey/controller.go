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

package vaulttransitkey

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

// Reconciler reconciles a VaultTransitKey object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	VaultAPI    vault.API
	Recorder    record.EventRecorder
	EventFilter predicate.Predicate
}

// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitkeys,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitkeys/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaulttransitkeys/finalizers,verbs=update

// Reconcile sets up the controller with the Manager.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaulttransitkey", req.NamespacedName)
	log.Info("reconciling for transit key")

	key := &heistv1alpha1.VaultTransitKey{}
	if err := r.Get(ctx, req.NamespacedName, key); err != nil {
		if err2 := client.IgnoreNotFound(err); err2 != nil {
			log.Error(err, "unable to fetch VaultTransitKey")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previous := key.DeepCopy()

	setDefaultConditions(key)

	result, err := r.performReconciliation(ctx, key)

	if deep.Equal(previous.Status, key.Status) != nil {
		if err := r.Status().Update(ctx, key); err != nil {
			return common.Requeue, err
		}
	}

	if deep.Equal(previous.Finalizers, key.Finalizers) != nil {
		if err := r.Update(ctx, key); err != nil {
			return common.Requeue, err
		}
	}

	return result, err
}

func (r *Reconciler) performReconciliation(ctx context.Context, key *heistv1alpha1.VaultTransitKey) (ctrl.Result, error) {
	if key.DeletionTimestamp != nil {
		return r.finalizeTransitKey(ctx, key)
	}

	return r.updateTransitKey(ctx, key)
}

func setDefaultConditions(key *heistv1alpha1.VaultTransitKey) {
	if meta.FindStatusCondition(key.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&key.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) getEngineForKey(ctx context.Context, key *heistv1alpha1.VaultTransitKey) (engine *heistv1alpha1.VaultTransitEngine, err error) {
	engine = &heistv1alpha1.VaultTransitEngine{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: key.Namespace,
			Name:      key.Spec.Engine,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(engine), engine); err != nil {
		return nil, err
	}
	return engine, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultTransitKey{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Watches(
			&heistv1alpha1.VaultTransitEngine{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, object client.Object) []reconcile.Request {
				engine, ok := object.(*heistv1alpha1.VaultTransitEngine)
				if !ok {
					return nil
				}

				transitKeys := heistv1alpha1.VaultTransitKeyList{}
				for i := 0; i < 3; i++ {
					if err := mgr.GetClient().List(context.TODO(), &transitKeys, &client.ListOptions{Namespace: engine.Namespace}); err != nil {
						time.Sleep(time.Second)
						continue
					}
					requests := make([]reconcile.Request, 0, len(transitKeys.Items))
					for _, secret := range transitKeys.Items {
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
