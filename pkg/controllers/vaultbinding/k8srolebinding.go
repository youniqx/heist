package vaultbinding

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateK8sRoleBinding(ctx context.Context, info *BindingInfo) error {
	binding := &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      info.K8sRoleName,
			Namespace: info.Binding.Namespace,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, binding, func() error {
		_ = controllerutil.SetControllerReference(info.Binding, binding, r.Scheme)

		binding.RoleRef = v1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     info.K8sRoleName,
		}

		binding.Subjects = []v1.Subject{
			{
				Kind:      v1.ServiceAccountKind,
				Name:      info.Spec.Subject.Name,
				Namespace: info.Binding.Namespace,
			},
		}

		return nil
	})
	if err != nil {
		return err
	}

	switch result {
	case controllerutil.OperationResultNone:
	case controllerutil.OperationResultCreated:
		r.Recorder.Eventf(info.Binding, "Normal", "RoleBinding", "RoleBinding %s has been created", binding.Name)
	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(info.Binding, "Normal", "RoleBinding", "RoleBinding %s has been updated", binding.Name)
	case controllerutil.OperationResultUpdatedStatus:
	case controllerutil.OperationResultUpdatedStatusOnly:
	}

	return nil
}
