package vaulttransitkey

import (
	"path/filepath"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/vault/policy"
)

func (r *Reconciler) updatePolicy(engine *heistv1alpha1.VaultTransitEngine, key *heistv1alpha1.VaultTransitKey) error {
	issuerPath, err := engine.GetMountPath()
	if err != nil {
		return err
	}

	keyName, err := key.GetTransitKeyName()
	if err != nil {
		return err
	}

	readPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyRead(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "keys", keyName),
				Capabilities: []policy.Capability{
					policy.ReadCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(readPolicy); err != nil {
		return err
	}

	encryptPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyEncrypt(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "encrypt", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(encryptPolicy); err != nil {
		return err
	}

	decryptPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyDecrypt(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "decrypt", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(decryptPolicy); err != nil {
		return err
	}

	rewrapPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyRewrap(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "rewrap", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(rewrapPolicy); err != nil {
		return err
	}

	dataKeyPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyDatakey(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "datakey", "plaintext", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "datakey", "wrapped", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(dataKeyPolicy); err != nil {
		return err
	}

	hmacPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyHmac(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "hmac", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "hmac", keyName, "sha2-224"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "hmac", keyName, "sha2-256"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "hmac", keyName, "sha2-384"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "hmac", keyName, "sha2-512"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(hmacPolicy); err != nil {
		return err
	}

	signPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeySign(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "sign", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "sign", keyName, "sha2-224"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "sign", keyName, "sha2-256"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "sign", keyName, "sha2-384"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "sign", keyName, "sha2-512"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	if err := r.VaultAPI.UpdatePolicy(signPolicy); err != nil {
		return err
	}

	verifyPolicy := &policy.Policy{
		Name: common.GetPolicyNameForTransitKeyVerify(key),
		Rules: []*policy.Rule{
			{
				Path: filepath.Join(issuerPath, "verify", keyName),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "verify", keyName, "sha2-224"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "verify", keyName, "sha2-256"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "verify", keyName, "sha2-384"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
			{
				Path: filepath.Join(issuerPath, "verify", keyName, "sha2-512"),
				Capabilities: []policy.Capability{
					policy.UpdateCapability,
				},
			},
		},
	}

	return r.VaultAPI.UpdatePolicy(verifyPolicy)
}
