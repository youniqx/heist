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
var vaultcertificateauthoritylog = logf.Log.WithName("vaultcertificateauthority-resource")

func (in *VaultCertificateAuthority) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(in).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-heist-youniqx-com-v1alpha1-vaultcertificateauthority,mutating=false,failurePolicy=fail,sideEffects=None,groups=heist.youniqx.com,resources=vaultcertificateauthorities,verbs=create;update;delete,versions=v1alpha1,name=vvaultcertificateauthority.heist.youniqx.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &VaultCertificateAuthority{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (in *VaultCertificateAuthority) ValidateCreate() error {
	log := vaultcertificateauthoritylog.WithName("validate").WithValues(
		"action", "create",
		"name", in.Name,
		"namespace", in.Namespace,
	)
	log.Info("create validation started")
	return in.validate(log)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (in *VaultCertificateAuthority) ValidateUpdate(old runtime.Object) error {
	log := vaultcertificateauthoritylog.WithName("validate").WithValues(
		"action", "update",
		"name", in.Name,
		"namespace", in.Namespace,
	)
	log.Info("update validation started")
	return in.validate(log)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (in *VaultCertificateAuthority) ValidateDelete() error {
	log := vaultcertificateauthoritylog.WithName("validate").WithValues(
		"action", "delete",
		"name", in.Name,
		"namespace", in.Namespace,
	)
	log.Info("delete validation started")

	if in.Spec.DeleteProtection {
		log.Info("rejecting change: resource has delete protection enabled. It cannot be deleted.")
		return errors.New("delete protection is enabled for this VaultCertificateAuthority, it cannot be deleted")
	}

	return nil
}

func (in *VaultCertificateAuthority) validate(log logr.Logger) error {
	if err := in.validateImportedCert(log); err != nil {
		return err
	}

	if err := in.validateCertSettings(log); err != nil {
		return err
	}

	return nil
}

func (in *VaultCertificateAuthority) validateCertSettings(log logr.Logger) error {
	if in.Spec.Import != nil {
		return nil
	}

	if in.Spec.Settings.KeyType == "" {
		log.Info("rejecting change: key_type is not set")
		return errors.New("key_type is not set")
	}

	if in.Spec.Settings.KeyBits == 0 {
		log.Info("rejecting change: key_bits is not set")
		return errors.New("key_bits is not set")
	}

	return nil
}

func (in *VaultCertificateAuthority) validateImportedCert(log logr.Logger) error {
	if in.Spec.Import == nil {
		return nil
	}

	if in.Spec.Import.PrivateKey == "" {
		log.Info("rejecting change: private key to import is not set.")
		return errors.New("private key to import is not set")
	}
	if !cipherTextRegex.MatchString(in.Spec.Import.PrivateKey) {
		log.Info("rejecting change: private key is not a valid encrypted string")
		return errors.New("private key is not a valid encrypted string")
	}

	if in.Spec.Import.Certificate == "" {
		log.Info("rejecting change: certificate to import is not set.")
		return errors.New("certificate to import is not set")
	}
	if !cipherTextRegex.MatchString(in.Spec.Import.PrivateKey) {
		log.Info("rejecting change: certificate is not a valid encrypted string")
		return errors.New("certificate is not a valid encrypted string")
	}

	return nil
}
