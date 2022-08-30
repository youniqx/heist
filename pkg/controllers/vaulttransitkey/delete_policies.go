package vaulttransitkey

import (
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/core"
)

func (r *Reconciler) deletePolicy(key *heistv1alpha1.VaultTransitKey) error {
	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyRead(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyEncrypt(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyDecrypt(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyRewrap(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyDatakey(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyHmac(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeySign(key))); err != nil {
		return err
	}

	if err := r.VaultAPI.DeletePolicy(core.PolicyName(common.GetPolicyNameForTransitKeyVerify(key))); err != nil {
		return err
	}

	return nil
}
