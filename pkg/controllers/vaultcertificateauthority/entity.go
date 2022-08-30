package vaultcertificateauthority

import (
	"time"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
	"github.com/youniqx/heist/pkg/vault/pki"
)

func (r *Reconciler) toVaultCAEntity(ca *heistv1alpha1.VaultCertificateAuthority) (pki.CAEntity, error) {
	path, err := ca.GetMountPath()
	if err != nil {
		return nil, err
	}

	var importedCert *pki.ImportedCert
	if ca.Spec.Import != nil {
		certificateBytes, err := r.VaultAPI.TransitDecrypt(managed.TransitEngine, managed.TransitKey, ca.Spec.Import.Certificate)
		if err != nil {
			return nil, err
		}

		privateKeyBytes, err := r.VaultAPI.TransitDecrypt(managed.TransitEngine, managed.TransitKey, ca.Spec.Import.PrivateKey)
		if err != nil {
			return nil, err
		}

		importedCert = &pki.ImportedCert{
			PrivateKey:  string(privateKeyBytes),
			Certificate: string(certificateBytes),
		}
	}

	var maxLeaseTTL time.Duration
	if ca.Spec.Settings.TTL.Duration != 0 && ca.Spec.Tuning.MaxLeaseTTL.Duration == 0 {
		maxLeaseTTL = ca.Spec.Settings.TTL.Duration
	} else {
		maxLeaseTTL = ca.Spec.Tuning.MaxLeaseTTL.Duration
	}

	return &pki.CA{
		PluginName: ca.Spec.Plugin,
		Path:       path,
		Settings: &pki.CASettings{
			SubjectAlternativeNames: ca.Spec.Settings.SubjectAlternativeNames,
			IPSans:                  ca.Spec.Settings.IPSans,
			URISans:                 ca.Spec.Settings.URISans,
			OtherSans:               ca.Spec.Settings.OtherSans,
			TTL:                     core.NewTTL(ca.Spec.Settings.TTL.Duration),
			KeyType:                 ca.Spec.Settings.KeyType,
			KeyBits:                 ca.Spec.Settings.KeyBits,
			ExcludeCNFromSans:       ca.Spec.Settings.ExcludeCNFromSans,
			PermittedDNSDomains:     ca.Spec.Settings.PermittedDNSDomains,
		},
		Subject: &pki.Subject{
			SubjectSettings: &pki.SubjectSettings{
				Organization:       ca.Spec.Subject.Organization,
				OrganizationalUnit: ca.Spec.Subject.OrganizationalUnit,
				Country:            ca.Spec.Subject.Country,
				Locality:           ca.Spec.Subject.Locality,
				Province:           ca.Spec.Subject.Province,
				StreetAddress:      ca.Spec.Subject.StreetAddress,
				PostalCode:         ca.Spec.Subject.PostalCode,
			},
			CommonName:   ca.Spec.Subject.CommonName,
			SerialNumber: "",
		},
		Config: &mount.TuneConfig{
			DefaultLeaseTTL: core.NewTTL(ca.Spec.Tuning.DefaultLeaseTTL.Duration),
			MaxLeaseTTL:     core.NewTTL(maxLeaseTTL),
			Description:     ca.Spec.Tuning.Description,
		},
		ImportedCert: importedCert,
	}, nil
}
