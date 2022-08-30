package vaultcertificaterole

import (
	"context"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/core"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) finalizeCertificate(ctx context.Context, certificate *heistv1alpha1.VaultCertificateRole) (ctrl.Result, error) {
	r.Recorder.Event(certificate, "Normal", "DeletionInProgress", "The VaultCertificateRole has been marked for deletion. CertificateChain will be deleted from Vault.")

	meta.SetStatusCondition(&certificate.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.Terminating,
		Message: "CertificateRole is in the process of being deleted from Vault",
	})

	if controllerutil.ContainsFinalizer(certificate, common.YouniqxFinalizer) {
		if err := r.deleteCertificate(ctx, certificate); err != nil {
			return common.Requeue, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) deleteCertificate(ctx context.Context, cert *heistv1alpha1.VaultCertificateRole) error {
	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForCertificateIssuing(cert))); err != nil {
		r.Recorder.Event(cert, "Warning", "ErrorDuringCertDeletion", "Failed to delete certificate policies from Vault")
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: "Failed to delete policies",
		})
		return err
	}

	issuer := &heistv1alpha1.VaultCertificateAuthority{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cert.Spec.Issuer,
			Namespace: cert.Namespace,
		},
	}

	if err := r.VaultAPI.DeleteCertificateRole(issuer, cert); err != nil {
		r.Recorder.Event(cert, "Warning", "ErrorDuringCertDeletion", "Failed to delete certificate role from Vault")
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorVault,
			Message: "Failed to delete certificate role",
		})
		return err
	}

	controllerutil.RemoveFinalizer(cert, common.YouniqxFinalizer)
	if err := r.Update(ctx, cert); err != nil {
		r.Recorder.Eventf(cert, "Warning", "RemoveFinalizerFailed", "Failed to remove finalizer from ca %s", cert.Name)
		meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
			Type:    heistv1alpha1.Conditions.Types.Provisioned,
			Status:  metav1.ConditionFalse,
			Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
			Message: "Failed to remove finalizer",
		})
		return client.IgnoreNotFound(err)
	}

	r.Recorder.Eventf(cert, "Normal", "FinalizerRemoved", "VaultCertificateAuthority %s is now ready for deletion from Kubernetes.", cert.Name)

	meta.SetStatusCondition(&cert.Status.Conditions, metav1.Condition{
		Type:    heistv1alpha1.Conditions.Types.Provisioned,
		Status:  metav1.ConditionFalse,
		Reason:  heistv1alpha1.Conditions.Reasons.ErrorKubernetes,
		Message: "deletion complete",
	})

	return nil
}
