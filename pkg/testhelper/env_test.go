package testhelper

import (
	"fmt"
	"testing"

	"github.com/go-test/deep"
)

func Test_getPluralNaming(t *testing.T) {
	type args struct {
		singularName string
	}
	tests := []struct {
		name           string
		args           args
		wantPluralName string
		wantErr        error
	}{
		{
			name: "should error on random singular",
			args: args{
				singularName: "some-random-crd",
			},
			wantErr: fmt.Errorf("plural name of resource some-random-crd cannot be resolved: %w", ErrUnknownResource),
		},
		{
			name:           "should convert VaultBinding",
			args:           args{singularName: "VaultBinding"},
			wantPluralName: "vaultbindings",
		},
		{
			name:           "should convert VaultCertificateRole",
			args:           args{singularName: "VaultCertificateRole"},
			wantPluralName: "vaultcertificateroles",
		},
		{
			name:           "should convert VaultClientConfig",
			args:           args{singularName: "VaultClientConfig"},
			wantPluralName: "vaultclientconfigs",
		},
		{
			name:           "should convert VaultKVSecretEngine",
			args:           args{singularName: "VaultKVSecretEngine"},
			wantPluralName: "vaultkvsecretengines",
		},
		{
			name:           "should convert VaultKVSecret",
			args:           args{singularName: "VaultKVSecret"},
			wantPluralName: "vaultkvsecrets",
		},
		{
			name:           "should convert VaultSyncSecret",
			args:           args{singularName: "VaultSyncSecret"},
			wantPluralName: "vaultsyncsecrets",
		},
		{
			name:           "should convert VaultTransitEngine",
			args:           args{singularName: "VaultTransitEngine"},
			wantPluralName: "vaulttransitengines",
		},
		{
			name:           "should convert VaultTransitKey",
			args:           args{singularName: "VaultTransitKey"},
			wantPluralName: "vaulttransitkeys",
		},
		{
			name:           "should convert VaultCertificateAuthority",
			args:           args{singularName: "VaultCertificateAuthority"},
			wantPluralName: "vaultcertificateauthorities",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPluralName, gotErr := getPluralNaming(tt.args.singularName)
			if diff := deep.Equal(tt.wantErr, gotErr); diff != nil {
				t.Errorf("getPluralNaming() = wantErr != gotErr: %s", diff)
			}

			if diff := deep.Equal(tt.wantPluralName, gotPluralName); diff != nil {
				t.Errorf("getPluralNaming() = wantPluralName != gotPluralName: %s", diff)
			}
		})
	}
}
