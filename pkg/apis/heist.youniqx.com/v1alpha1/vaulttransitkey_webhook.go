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

package v1alpha1

import (
	"errors"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var vaulttransitkeylog = logf.Log.WithName("vaulttransitkey-resource")

func (r *VaultTransitKey) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-heist-youniqx-com-v1alpha1-vaulttransitkey,mutating=false,failurePolicy=fail,sideEffects=None,groups=heist.youniqx.com,resources=vaulttransitkeys,verbs=create;update;delete,versions=v1alpha1,name=vvaulttransitkey.heist.youniqx.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &VaultTransitKey{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *VaultTransitKey) ValidateCreate() error {
	log := vaulttransitkeylog.WithName("validate").WithValues(
		"action", "create",
		"name", r.Name,
		"namespace", r.Namespace,
	)
	log.Info("create validation started")
	return r.validate(log)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *VaultTransitKey) ValidateUpdate(old runtime.Object) error {
	log := vaulttransitkeylog.WithName("validate").WithValues(
		"action", "update",
		"name", r.Name,
		"namespace", r.Namespace,
	)
	log.Info("update validation started")
	return r.validate(log)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *VaultTransitKey) ValidateDelete() error {
	log := vaulttransitkeylog.WithName("validate").WithValues(
		"action", "delete",
		"name", r.Name,
		"namespace", r.Namespace,
	)
	log.Info("delete validation started")

	if r.Spec.DeleteProtection {
		log.Info("rejecting change: resource has delete protection enabled. It cannot be deleted.")
		return errors.New("delete protection is enabled for this VaultTransitKey, it cannot be deleted")
	}

	return nil
}

func (r *VaultTransitKey) validate(_ logr.Logger) error {
	return nil
}
