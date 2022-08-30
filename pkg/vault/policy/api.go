package policy

import "github.com/youniqx/heist/pkg/vault/core"

type policyAPI struct {
	Core core.API
}

func NewAPI(core core.API) API {
	return &policyAPI{Core: core}
}

type API interface {
	UpdatePolicy(policy Entity) error
	DeletePolicy(policy core.PolicyNameEntity) error
	ReadPolicy(policy core.PolicyNameEntity) (*Policy, error)
}

type Entity interface {
	core.PolicyNameEntity
	Body
}

type Body interface {
	GetPolicyRules() ([]*Rule, error)
}
