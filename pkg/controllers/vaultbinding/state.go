package vaultbinding

import (
	"context"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) findDominantBinding(ctx context.Context, binding *heistv1alpha1.VaultBinding, spec *heistv1alpha1.VaultBindingSpec) (*heistv1alpha1.VaultBinding, error) {
	bindings := &heistv1alpha1.VaultBindingList{}
	if err := r.List(ctx, bindings, client.InNamespace(binding.Namespace)); err != nil {
		return nil, err
	}

	var rivals []heistv1alpha1.VaultBinding

	for _, item := range bindings.Items {
		itemSubject := item.Spec.Subject
		if itemSubject.Name == spec.Subject.Name {
			rivals = append(rivals, item)
		}
	}

	dominantBinding := binding

	for _, rival := range rivals {
		if rival.DeletionTimestamp != nil {
			continue
		}
		if rival.CreationTimestamp.Before(&dominantBinding.CreationTimestamp) {
			dominantBinding = rival.DeepCopy()
		}
	}

	return dominantBinding, nil
}
