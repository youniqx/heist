package vaultbinding

import (
	"context"

	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *Reconciler) updateK8sRole(ctx context.Context, info *BindingInfo) error {
	role := &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      info.K8sRoleName,
			Namespace: info.Binding.Namespace,
		},
	}

	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, role, func() error {
		_ = controllerutil.SetControllerReference(info.Binding, role, r.Scheme)

		role.Rules = []v1.PolicyRule{
			{
				Verbs:         []string{"get", "list", "watch"},
				APIGroups:     []string{"heist.youniqx.com"},
				Resources:     []string{"vaultclientconfigs"},
				ResourceNames: []string{},
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
		r.Recorder.Eventf(info.Binding, "Normal", "Role", "Role %s has been created", role.Name)
	case controllerutil.OperationResultUpdated:
		r.Recorder.Eventf(info.Binding, "Normal", "Role", "Role %s has been updated", role.Name)
	case controllerutil.OperationResultUpdatedStatus:
	case controllerutil.OperationResultUpdatedStatusOnly:
	}

	return nil
}
