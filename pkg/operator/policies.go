package operator

import (
	"github.com/youniqx/heist/pkg/vault/policy"
)

var operatorPolicyRules = []*policy.Rule{
	{
		Path: "managed/*",
		Capabilities: []policy.Capability{
			policy.CreateCapability,
			policy.ReadCapability,
			policy.UpdateCapability,
			policy.DeleteCapability,
			policy.ListCapability,
		},
	},
	{
		Path: "sys/mounts",
		Capabilities: []policy.Capability{
			policy.ReadCapability,
		},
	},
	{
		Path: "sys/plugins/reload/backend",
		Capabilities: []policy.Capability{
			policy.UpdateCapability,
		},
	},
	{
		Path: "sys/tools/random/*",
		Capabilities: []policy.Capability{
			policy.UpdateCapability,
		},
	},
	{
		Path: "sys/policies/acl/*",
		Capabilities: []policy.Capability{
			policy.CreateCapability,
			policy.ReadCapability,
			policy.UpdateCapability,
			policy.DeleteCapability,
		},
	},
	{
		Path: "sys/mounts/managed/*",
		Capabilities: []policy.Capability{
			policy.CreateCapability,
			policy.ReadCapability,
			policy.UpdateCapability,
			policy.DeleteCapability,
			policy.ListCapability,
		},
	},
	{
		Path: "auth/managed/kubernetes/role/*",
		Capabilities: []policy.Capability{
			policy.CreateCapability,
			policy.ReadCapability,
			policy.UpdateCapability,
			policy.DeleteCapability,
		},
	},
}
