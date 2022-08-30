package vaultcertificaterole

import (
	"context"
	"path/filepath"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/policy"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateCertificate(ctx context.Context, cert *heistv1alpha1.VaultCertificateRole) (ctrl.Result, error) {
	if err := r.attachFinalizer(ctx, cert); err != nil {
		r.Recorder.Eventf(cert, "Warning", "AttachFinalizerFailed", "Failed to attach the finalizer to cert %s", cert.Name)
		r.Log.Error(err, "unable to attach finalizer")
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
			Message: "Finalizer could not be attached to this VaultCertificateRole",
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	issuer := &heistv1alpha1.VaultCertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cert.Spec.Issuer,
			Namespace: cert.Namespace,
		},
	}

	if err := r.Get(ctx, client.ObjectKeyFromObject(issuer), issuer); err != nil {
		r.Recorder.Eventf(cert, "Warning", "ErrorConfig", "Issuer %s specified in certificate %s could not be found", cert.Spec.Issuer, cert.Name)
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorConfig,
			Message: "Referenced issuer could not be found",
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if err := r.VaultAPI.UpdateCertificateRole(issuer, cert); err != nil {
		r.Recorder.Eventf(cert, "Warning", "CertificateRoleError", "Failed to update certificate role %s", cert.Name)
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: "CertificateRole could not be created",
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if err := r.updatePolicy(issuer, cert); err != nil {
		r.Recorder.Eventf(cert, "Warning", "PolicyUpdateFailed", "Failed to creates policies for certificate role %s", cert.Name)
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: "Policies for certificate role could not be created",
		})
		return common.Requeue, client.IgnoreNotFound(err)
	}

	if meta.IsStatusConditionFalse(cert.Status.Conditions, heistv1alpha1.Conditions.Types.Provisioned) {
		r.Recorder.Eventf(cert, "Normal", "ProvisioningSuccessful", "CertificateRole %s has been provisioned", cert.Name)
	}

	meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionTrue,
		Reason:  heistv1alpha1.Conditions.Reasons.Provisioned,
		Message: "CertificateRole has been provisioned",
	})

	return ctrl.Result{}, nil
}

func (r *Reconciler) updatePolicy(issuer *heistv1alpha1.VaultCertificateAuthority, cert *heistv1alpha1.VaultCertificateRole) error {
	issuerPath, err := issuer.GetMountPath()
	if err != nil {
		return err
	}

	roleName, err := cert.GetRoleName()
	if err != nil {
		return err
	}

	issuePolicy := &policy.Policy{
		Name: common.GetPolicyNameForCertificateIssuing(cert),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "issue", roleName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(issuePolicy); err != nil {
		return err
	}

	signCsrPolicy := &policy.Policy{
		Name: common.GetPolicyNameForCertificateSignCSR(cert),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "sign", roleName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(signCsrPolicy); err != nil {
		return err
	}

	signVerbatimPolicy := &policy.Policy{
		Name: common.GetPolicyNameForCertificateSignVerbatim(cert),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "sign-verbatim", roleName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	return r.VaultAPI.UpdatePolicy(signVerbatimPolicy)
}

func (r *Reconciler) attachFinalizer(ctx context.Context, certificate *heistv1alpha1.VaultCertificateRole) error {
	if !controllerutil.ContainsFinalizer(certificate, common.YouniqxFinalizer) {
		controllerutil.AddFinalizer(certificate, common.YouniqxFinalizer)
		if err := r.Update(ctx, certificate); err != nil {
			return err
		}
		r.Recorder.Event(certificate, "Normal", "FinalizerAttached", "The finalizer has been attached to this object")
	}

	return nil
}
