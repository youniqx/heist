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

package vaultbinding

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

const fileModeBase = 8

// Reconciler reconciles a VaultBinding object.
type Reconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Recorder    record.EventRecorder
	VaultAPI    vault.API
	EventFilter predicate.Predicate
}

// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultbindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultbindings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=heist.youniqx.com,resources=vaultbindings/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=roles;rolebindings,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VaultBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("vaultbinding", req.NamespacedName)
	log.Info("reconciling for binding")

	binding := &heistv1alpha1.VaultBinding{}
	if err := r.Get(ctx, req.NamespacedName, binding); err != nil {
		if err := client.IgnoreNotFound(err); err != nil {
			log.Info("unable to fetch VaultBinding", "error", err)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	previous := binding.DeepCopy()

	setDefaultConditions(binding)

	result, err := r.performReconciliation(ctx, binding, buildEffectiveSpec(binding))

	if deep.Equal(previous.Status, binding.Status) != nil {
		if err := r.Status().Update(ctx, binding); err != nil {
			return common.Requeue, err
		}
	}

	if deep.Equal(previous.Finalizers, binding.Finalizers) != nil {
		if err := r.Update(ctx, binding); err != nil {
			return common.Requeue, err
		}
	}

	return result, err
}

func setDefaultConditions(binding *heistv1alpha1.VaultBinding) {
	if meta.FindStatusCondition(binding.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) == nil {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
	if meta.FindStatusCondition(binding.Status.Conditions, heistv1alpha1.Conditions.Types.Active) == nil {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Active,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.Initializing,
			Message: "provisioning is about to start",
		})
	}
}

func (r *Reconciler) performReconciliation(ctx context.Context, binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) (ctrl.Result, error) {
	if binding.GetDeletionTimestamp() != nil {
		return r.finalizeBinding(binding, spec)
	}

	binding.Status.AppliedSpec = *spec

	return r.updateBinding(ctx, binding, spec)
}

//nolint:gocognit
func buildEffectiveSpec(binding *heistv1alpha1.VaultBinding) *heistv1alpha1.VaultBindingSpec {
	spec := binding.Spec.DeepCopy()

	setDefaultTemplateModes(spec)
	setDefaultCapabilities(spec)

	return spec
}

func setDefaultTemplateModes(spec *heistv1alpha1.VaultBindingSpec) {
	for index, template := range spec.Agent.Templates {
		if template.Mode == "" {
			spec.Agent.Templates[index].Mode = "0640"
		}
	}
}

func setDefaultCapabilities(spec *heistv1alpha1.VaultBindingSpec) {
	for index, secret := range spec.KVSecrets {
		if len(secret.Capabilities) == 0 {
			spec.KVSecrets[index].Capabilities = []heistv1alpha1.VaultBindingKVCapability{
				heistv1alpha1.VaultBindingKVCapabilityRead,
			}
		}
	}

	for index, cert := range spec.CertificateRoles {
		if len(cert.Capabilities) == 0 {
			spec.CertificateRoles[index].Capabilities = []heistv1alpha1.VaultBindingCertificateCapability{
				heistv1alpha1.VaultBindingCertificateCapabilityIssue,
			}
		}
	}

	for index, ca := range spec.CertificateAuthorities {
		if len(ca.Capabilities) == 0 {
			spec.CertificateAuthorities[index].Capabilities = []heistv1alpha1.VaultBindingCertificateAuthorityCapability{
				heistv1alpha1.VaultBindingCertificateAuthorityCapabilityReadPublic,
			}
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&heistv1alpha1.VaultBinding{}).
		WithEventFilter(r.EventFilter).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}
