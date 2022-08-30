package policy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/youniqx/heist/pkg/vault/core"
)

var (
	// CreateCapability gives permission to create objects.
	CreateCapability Capability = "create"
	// ReadCapability gives permission to read objects.
	ReadCapability Capability = "read"
	// UpdateCapability gives permission to update objects.
	UpdateCapability Capability = "update"
	// DeleteCapability gives permission to delete objects.
	DeleteCapability Capability = "delete"
	// ListCapability gives permission to list objects.
	ListCapability Capability = "list"
)

type Capability string

type Rule struct {
	Path         string       `hcl:"name,label"`
	Capabilities []Capability `hcl:"capabilities"`
}

type Policy struct {
	Name  string
	Rules []*Rule `hcl:"path,block"`
}

func (p *Policy) GetPolicyName() (string, error) {
	return p.Name, nil
}

func (p *Policy) GetPolicyRules() ([]*Rule, error) {
	return p.Rules, nil
}

func (p *Policy) MarshalJSON() ([]byte, error) {
	hcl := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(p, hcl.Body())

	policyString := string(hcl.Bytes())
	policyString = strings.TrimSpace(policyString)

	data, err := json.Marshal(policyString)
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("failed to encode policy hcl to json").WithCause(err)
	}

	return data, nil
}

func (p *Policy) UnmarshalJSON(bytes []byte) error {
	var policy string
	if err := json.Unmarshal(bytes, &policy); err != nil {
		return core.ErrAPIError.WithDetails("failed to decode policy hcl from json").WithCause(err)
	}

	parser := hclparse.NewParser()
	hcl, diagnostics := parser.ParseHCL([]byte(policy), "policy.hcl")

	if diagnostics != nil {
		return diagnostics
	}

	if err := gohcl.DecodeBody(hcl.Body, nil, p); err != nil {
		return fmt.Errorf("failed to decode hcl: %s: %w", string(bytes), err)
	}

	return nil
}
