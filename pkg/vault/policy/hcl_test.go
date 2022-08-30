package policy

import (
	"reflect"
	"testing"

	"github.com/go-test/deep"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

func TestEncodeHCL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		policy *Policy
		want   string
	}{
		{
			name: "should correctly encode single policy",
			policy: &Policy{
				Name: "policy-name",
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
				},
			},
			want: "\npath \"test-policy\" {\n  capabilities = [\"create\", \"update\"]\n}\n",
		},
		{
			name: "should correctly encode multiple policies",
			policy: &Policy{
				Name: "policy-name",
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
					{
						Path: "test-policy-2",
						Capabilities: []Capability{
							CreateCapability,
							ReadCapability,
							UpdateCapability,
							DeleteCapability,
							ListCapability,
						},
					},
				},
			},
			want: "\npath \"test-policy\" {\n  capabilities = [\"create\", \"update\"]\n}\npath \"test-policy-2\" {\n  capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f := hclwrite.NewEmptyFile()
			gohcl.EncodeIntoBody(tt.policy, f.Body())
			got := string(f.Bytes())
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeHCL Policy = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestDecodeHCL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		stringPolicy string
		wantErr      bool
		want         *Policy
	}{
		{
			name:         "should correctly decode single policy",
			stringPolicy: "\npath \"test-policy\" {\n  capabilities = [\"create\", \"update\"]\n}\n",
			want: &Policy{
				Name: "",
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
				},
			},
		},
		{
			name:         "should correctly encode multiple policies",
			stringPolicy: "\npath \"test-policy\" {\n capabilities = [\"create\", \"update\"]\n}\npath \"test-policy-2\" {\n capabilities = [\"create\", \"read\", \"update\", \"delete\", \"list\"]\n}\n",
			want: &Policy{
				Name: "",
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
					{
						Path: "test-policy-2",
						Capabilities: []Capability{
							CreateCapability,
							ReadCapability,
							UpdateCapability,
							DeleteCapability,
							ListCapability,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := &Policy{}
			parser := hclparse.NewParser()
			f, err := parser.ParseHCL([]byte(tt.stringPolicy), "test-case-1")
			if err != nil {
				t.Errorf("unexpected error during hcl parsing")
				return
			}

			err = gohcl.DecodeBody(f.Body, nil, got)
			if err != nil != tt.wantErr {
				t.Errorf("wantErr = %t, got %v", tt.wantErr, err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EncodeHCL Policy = \"%v\", want \"%v\"", got, tt.want)
			}
		})
	}
}

func TestPolicyRoot_MarshalJSON(t *testing.T) {
	t.Parallel()

	type fields struct {
		Rules []*Rule
	}

	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "should correctly encode single policy",
			fields: fields{
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
				},
			},
			want:    []byte("\"path \\\"test-policy\\\" {\\n  capabilities = [\\\"create\\\", \\\"update\\\"]\\n}\""),
			wantErr: false,
		},
		{
			name: "should correctly encode multiple policies",
			fields: fields{
				Rules: []*Rule{
					{
						Path: "test-policy",
						Capabilities: []Capability{
							CreateCapability,
							UpdateCapability,
						},
					},
					{
						Path: "test-policy-2",
						Capabilities: []Capability{
							CreateCapability,
							ReadCapability,
							UpdateCapability,
							DeleteCapability,
							ListCapability,
						},
					},
				},
			},
			want:    []byte("\"path \\\"test-policy\\\" {\\n  capabilities = [\\\"create\\\", \\\"update\\\"]\\n}\\npath \\\"test-policy-2\\\" {\\n  capabilities = [\\\"create\\\", \\\"read\\\", \\\"update\\\", \\\"delete\\\", \\\"list\\\"]\\n}\""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &Policy{
				Rules: tt.fields.Rules,
			}
			got, err := p.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPolicyRoot_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	type args struct {
		bytes []byte
	}

	tests := []struct {
		name    string
		args    args
		want    []*Rule
		wantErr bool
	}{
		{
			name: "should correctly decode single policy",
			args: args{
				bytes: []byte("\"path \\\"test-policy\\\" {\\n  capabilities = [\\\"create\\\", \\\"update\\\"]\\n}\""),
			},
			want: []*Rule{
				{
					Path: "test-policy",
					Capabilities: []Capability{
						CreateCapability,
						UpdateCapability,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should correctly encode multiple policies",
			args: args{
				bytes: []byte("\"path \\\"test-policy\\\" {\\n  capabilities = [\\\"create\\\", \\\"update\\\"]\\n}\\npath \\\"test-policy-2\\\" {\\n  capabilities = [\\\"create\\\", \\\"read\\\", \\\"update\\\", \\\"delete\\\", \\\"list\\\"]\\n}\""),
			},
			want: []*Rule{
				{
					Path: "test-policy",
					Capabilities: []Capability{
						CreateCapability,
						UpdateCapability,
					},
				},
				{
					Path: "test-policy-2",
					Capabilities: []Capability{
						CreateCapability,
						ReadCapability,
						UpdateCapability,
						DeleteCapability,
						ListCapability,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := &Policy{}
			if err := got.UnmarshalJSON(tt.args.bytes); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := deep.Equal(got.Rules, tt.want); diff != nil {
				t.Errorf("UnmarshalJSON() diff = %v", diff)
			}
		})
	}
}
