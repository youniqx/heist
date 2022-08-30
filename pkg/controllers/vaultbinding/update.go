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
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateBinding(ctx context.Context, binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) (ctrl.Result, error) {
	r.attachFinalizer(binding)

	dominantBinding, err := r.findDominantBinding(ctx, binding, spec)
	if err != nil {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Failed to check for rival status: %v", err),
		})
		return common.Requeue, err
	}

	if dominantBinding != binding {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  "inactive",
			Message: fmt.Sprintf("Another binding is active for subject %s: %s ", spec.Subject.Name, dominantBinding.Name),
		})
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Active,
			Status:  metav1.ConditionFalse,
			Reason:  "rival_active",
			Message: fmt.Sprintf("Another binding is active for subject %s: %s ", spec.Subject.Name, dominantBinding.Name),
		})
		return common.Requeue, nil
	}

	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Active,
		Status:  metav1.ConditionTrue,
		Reason:  "active",
		Message: fmt.Sprintf("This binding is active for subject: %s", spec.Subject.Name),
	})

	bindingInfo, err := r.buildBindingInfo(ctx, binding, spec)
	if err != nil {
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: fmt.Sprintf("Failed to collect policy information: %v", err),
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if err := r.updateVaultRole(bindingInfo); err != nil {
		r.Recorder.Eventf(binding, "Normal", "ConfiguringRoleFailed", "Could not create role based on binding %s: %v", binding.Name, err)
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Could not create role: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.updateClientConfig(ctx, bindingInfo); err != nil {
		r.Recorder.Eventf(binding, "Normal", "ConfiguringClientConfigFailed", "Could not create client config based on binding %s: %v", binding.Name, err)
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Could not create client config: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.updateK8sRole(ctx, bindingInfo); err != nil {
		r.Recorder.Eventf(binding, "Normal", "ConfiguringRoleFailed", "Could not create K8s Role based on binding %s: %v", binding.Name, err)
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Could not create role: %v", err),
		})
		return common.Requeue, err
	}

	if err := r.updateK8sRoleBinding(ctx, bindingInfo); err != nil {
		r.Recorder.Eventf(binding, "Normal", "ConfiguringRoleBindingFailed", "Could not create K8s RoleBinding based on binding %s: %v", binding.Name, err)
		meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: fmt.Sprintf("Could not create role binding: %v", err),
		})
		return common.Requeue, err
	}

	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "Binding has been provisioned",
	})

	return ctrl.Result{}, nil
}

func (r *Reconciler) attachFinalizer(binding *heistv1alpha1.VaultBinding) {
	if controllerutil.ContainsFinalizer(binding, common.YouniqxFinalizer) {
		return
	}

	controllerutil.AddFinalizer(binding, common.YouniqxFinalizer)
	r.Recorder.Eventf(binding, "Normal", "FinalizerAttached", "Attached a finalizer to binding %s", binding.Name)
}
