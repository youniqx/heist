package agent

import (
	"testing"
)

func Test_agent_createOutputPath(t *testing.T) {
	type fields struct {
		BasePath string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "should not modify absolute paths",
			fields: fields{
				BasePath: "/heist",
			},
			args: args{path: "/vault/secrets/env/OIDCIDP_OP_AUTHZ_APIACCESSTOKEN"},
			want: "/vault/secrets/env/OIDCIDP_OP_AUTHZ_APIACCESSTOKEN",
		},
		{
			name: "should modify relative paths",
			fields: fields{
				BasePath: "/heist",
			},
			args: args{path: "env/OIDCIDP_OP_AUTHZ_APIACCESSTOKEN"},
			want: "/heist/secrets/env/OIDCIDP_OP_AUTHZ_APIACCESSTOKEN",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &agent{
				BasePath: tt.fields.BasePath,
			}
			if got := a.createOutputPath(tt.args.path); got != tt.want {
				t.Errorf("createOutputPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
