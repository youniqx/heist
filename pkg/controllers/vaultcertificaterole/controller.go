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

package vaultcertificaterole

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
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Reconciler reconciles a VaultCertificateRole object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	VaultAPI    vault.API
	Recorder    record.EventRecorder
	EventFilter predicate.Predicate
}

//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultcertificateroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultcertificateroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultcertificateroles/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaultcertificaterole", req.NamespacedName)
	log.Info("reconciling for vault certificate")

	certificate := &heistv1alpha1.VaultCertificateRole{}
	if err := r.Get(ctx, req.NamespacedName, certificate); err != nil {
		if err2 := client.IgnoreNotFound(err); err2 != nil {
			log.Error(err, "unable to fetch VaultCertificateRole")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previousStatus := certificate.Status.DeepCopy()

	setDefaultConditions(certificate)

	result, err := r.performReconciliation(ctx, certificate)

	if diff := deep.Equal(previousStatus, certificate.Status); diff != nil {
		if err := r.Status().Update(ctx, certificate); err != nil {
			result = common.Requeue
		}
	}

	return result, err
}

func setDefaultConditions(certificate *heistv1alpha1.VaultCertificateRole) {
	if meta.FindStatusCondition(certificate.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&certificate.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) performReconciliation(ctx context.Context, certificate *heistv1alpha1.VaultCertificateRole) (ctrl.Result, error) {
	if certificate.GetDeletionTimestamp() != nil {
		return r.finalizeCertificate(ctx, certificate)
	}

	return r.updateCertificate(ctx, certificate)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultCertificateRole{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Watches(&source.Kind{Type: &heistv1alpha1.VaultCertificateAuthority{}}, handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
			ca, ok := object.(*heistv1alpha1.VaultCertificateAuthority)
			if !ok {
				return nil
			}

			certificates := heistv1alpha1.VaultCertificateRoleList{}
			for i := 0; i < 3; i++ {
				if err := mgr.GetClient().List(context.TODO(), &certificates, &client.ListOptions{Namespace: ca.Namespace}); err != nil {
					time.Sleep(time.Second)
					continue
				}
				requests := make([]reconcile.Request, 0, len(certificates.Items))
				for _, certificate := range certificates.Items {
					if certificate.Spec.Issuer != ca.Name {
						continue
					}

					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name:      certificate.Name,
							Namespace: certificate.Namespace,
						},
					})
				}
				return requests
			}
			return nil
		})).
		Complete(r)
}
