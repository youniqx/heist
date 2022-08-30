package policy

import (
	"errors"
	"net/http"
	"path/filepath"

	"github.com/youniqx/heist/pkg/httpclient"
	"github.com/youniqx/heist/pkg/vault/core"
)

type getPolicyResponse struct {
	RequestID string        `json:"request_id"`
	Data      getPolicyData `json:"data"`
}

type getPolicyData struct {
	Name   string  `json:"name"`
	Policy *Policy `json:"policy"`
}

type setPolicyRequest struct {
	Policy *Policy `json:"policy"`
}

func (p *policyAPI) fetchPolicy(path string) (*getPolicyResponse, error) {
	log := p.Core.Log().WithValues("method", "fetchPolicy", "path", path)

	response := &getPolicyResponse{}
	if err := p.Core.MakeRequest(core.MethodGet, path, nil, httpclient.JSON(response, httpclient.ConstraintSuccess)); err != nil {
		log.Info("failed to fetch policy data", "error", err)

		var responseError *core.VaultHTTPError
		if errors.As(err, &responseError) && responseError.StatusCode == http.StatusNotFound {
			return nil, core.ErrDoesNotExist.WithCause(err)
		}

		return nil, core.ErrAPIError.WithDetails("failed to fetch policy data").WithCause(err)
	}

	return response, nil
}

func (p *policyAPI) deletePolicy(path string) error {
	log := p.Core.Log().WithValues("deletePolicy", path)

	if err := p.Core.MakeRequest(core.MethodDelete, path, nil, nil); err != nil {
		log.Info("failed to delete policy", "error", err)
		return core.ErrAPIError.WithDetails("failed to delete policy").WithCause(err)
	}

	return nil
}

func (p *policyAPI) writePolicy(path string, rules []*Rule) error {
	log := p.Core.Log().WithValues("writePolicy", path, "rules", rules)

	request := &setPolicyRequest{
		Policy: &Policy{
			Rules: rules,
		},
	}

	if err := p.Core.MakeRequest(core.MethodPost, path, httpclient.JSON(request), nil); err != nil {
		log.Info("couldn't write policy data", "path", path, "error", err)
		return core.ErrAPIError.WithDetails("failed to write policy data").WithCause(err)
	}

	return nil
}

func getPolicyPath(policy core.PolicyNameEntity) (string, error) {
	name, err := policy.GetPolicyName()
	if err != nil {
		return "", core.ErrAPIError.WithDetails("failed to fetch policy name").WithCause(err)
	}

	path := filepath.Join("/v1/sys/policies/acl/", name)

	return path, nil
}
