package vaulttransitkey

import (
	"testing"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault/transit"
)

func Test_hasChangedKeyType(t *testing.T) {
	type args struct {
		key *heistv1alpha1.VaultTransitKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Type: transit.TypeAes256Gcm96},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Type: transit.TypeRSA2048},
					},
				},
			},
			want: true,
		},
		{
			name: "should not detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Type: transit.TypeAes256Gcm96},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Type: transit.TypeAes256Gcm96},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasChangedKeyType(tt.args.key); got != tt.want {
				t.Errorf("hasChangedKeyType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasChangedEngine(t *testing.T) {
	type args struct {
		key *heistv1alpha1.VaultTransitKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Engine: "new-engine"},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Engine: "old-engine"},
					},
				},
			},
			want: true,
		},
		{
			name: "should not detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Engine: "old-engine"},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Engine: "old-engine"},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasChangedEngine(tt.args.key); got != tt.want {
				t.Errorf("hasChangedEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasChangedExportable(t *testing.T) {
	type args struct {
		key *heistv1alpha1.VaultTransitKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Exportable: false},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Exportable: true},
					},
				},
			},
			want: true,
		},
		{
			name: "should not detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{Exportable: true},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{Exportable: true},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasChangedExportable(tt.args.key); got != tt.want {
				t.Errorf("hasChangedExportable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasChangedAllowPlaintextBackup(t *testing.T) {
	type args struct {
		key *heistv1alpha1.VaultTransitKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{AllowPlaintextBackup: false},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{AllowPlaintextBackup: true},
					},
				},
			},
			want: true,
		},
		{
			name: "should not detect change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{AllowPlaintextBackup: true},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{AllowPlaintextBackup: true},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasChangedAllowPlaintextBackup(tt.args.key); got != tt.want {
				t.Errorf("hasChangedAllowPlaintextBackup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hasChangedKey(t *testing.T) {
	type args struct {
		key *heistv1alpha1.VaultTransitKey
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should return false on empty applied spec",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{
						Engine:               "some-engine",
						Type:                 transit.TypeRSA2048,
						Exportable:           true,
						AllowPlaintextBackup: true,
					},
					Status: heistv1alpha1.VaultTransitKeyStatus{},
				},
			},
			want: false,
		},
		{
			name: "should return false on no relevant change",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{
						Engine:               "some-engine",
						Type:                 transit.TypeRSA2048,
						Exportable:           true,
						AllowPlaintextBackup: true,
					},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{
							Engine:               "some-engine",
							Type:                 transit.TypeRSA2048,
							Exportable:           true,
							AllowPlaintextBackup: true,
						},
					},
				},
			},
			want: false,
		},
		{
			name: "should return true on incompatible changes",
			args: args{
				key: &heistv1alpha1.VaultTransitKey{
					Spec: heistv1alpha1.VaultTransitKeySpec{
						Engine:               "some-engine",
						Type:                 transit.TypeAes256Gcm96,
						Exportable:           true,
						AllowPlaintextBackup: true,
					},
					Status: heistv1alpha1.VaultTransitKeyStatus{
						AppliedSpec: heistv1alpha1.VaultTransitKeySpec{
							Engine:               "some-engine",
							Type:                 transit.TypeRSA2048,
							Exportable:           true,
							AllowPlaintextBackup: true,
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasIncompatibleChanges(tt.args.key); got != tt.want {
				t.Errorf("hasIncompatibleChanges() = %v, want %v", got, tt.want)
			}
		})
	}
}
