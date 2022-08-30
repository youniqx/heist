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
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) finalizeBinding(binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) (ctrl.Result, error) {
	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "Binding is being deleted",
	})
	meta.SetStatusCondition(&binding.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Active,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "Binding is being deleted",
	})

	if err := r.deleteVaultRole(binding, spec); err != nil {
		return common.Requeue, err
	}

	r.detachFinalizer(binding)

	return ctrl.Result{}, nil
}

func (r *Reconciler) detachFinalizer(binding *heistv1alpha1.VaultBinding) {
	if !controllerutil.ContainsFinalizer(binding, common.YouniqxFinalizer) {
		return
	}
	controllerutil.RemoveFinalizer(binding, common.YouniqxFinalizer)
	r.Recorder.Eventf(binding, "Normal", "FinalizerRemoved", "Finalizer has been removed from binding %s", binding.Name)
}
