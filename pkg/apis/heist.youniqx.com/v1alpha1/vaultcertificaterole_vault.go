package v1alpha1

import (
	"fmt"

	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/pki"
)

func (in *VaultCertificateRole) GetRoleName() (string, error) {
	return fmt.Sprintf("managed.pki.cert.%s.%s", in.Namespace, in.Name), nil
}

func (in *VaultCertificateRole) GetSettings() (*pki.RoleSettings, error) {
	return &pki.RoleSettings{
		TTL:                           core.NewTTL(in.Spec.Settings.TTL.Duration),
		MaxTTL:                        core.NewTTL(in.Spec.Settings.MaxTTL.Duration),
		AllowLocalhost:                in.Spec.Settings.AllowLocalhost,
		AllowedDomains:                in.Spec.Settings.AllowedDomains,
		AllowedDomainsTemplate:        in.Spec.Settings.AllowedDomainsTemplate,
		AllowBareDomains:              in.Spec.Settings.AllowBareDomains,
		AllowSubdomains:               in.Spec.Settings.AllowSubdomains,
		AllowGlobDomains:              in.Spec.Settings.AllowGlobDomains,
		AllowAnyName:                  in.Spec.Settings.AllowAnyName,
		EnforceHostNames:              in.Spec.Settings.EnforceHostNames,
		AllowIPSans:                   in.Spec.Settings.AllowIPSans,
		AllowedURISans:                in.Spec.Settings.AllowedURISans,
		AllowedOtherSans:              in.Spec.Settings.AllowedOtherSans,
		ServerFlag:                    in.Spec.Settings.ServerFlag,
		ClientFlag:                    in.Spec.Settings.ClientFlag,
		CodeSigningFlag:               in.Spec.Settings.CodeSigningFlag,
		EmailProtectionFlag:           in.Spec.Settings.EmailProtectionFlag,
		KeyType:                       in.Spec.Settings.KeyType,
		KeyBits:                       in.Spec.Settings.KeyBits,
		KeyUsage:                      in.Spec.Settings.KeyUsage,
		ExtendedKeyUsage:              in.Spec.Settings.ExtendedKeyUsage,
		ExtendedKeyUsageOids:          in.Spec.Settings.ExtendedKeyUsageOIDS,
		UseCSRCommonName:              in.Spec.Settings.UseCSRCommonName,
		UseCSRSans:                    in.Spec.Settings.UseCSRSans,
		GenerateLease:                 true,
		NoStore:                       false,
		RequireCommonName:             in.Spec.Settings.RequireCommonName,
		PolicyIdentifiers:             in.Spec.Settings.PolicyIdentifiers,
		BasicConstraintsValidForNonCA: in.Spec.Settings.BasicConstraintsValidForNonCA,
		NotBeforeDuration:             core.NewTTL(in.Spec.Settings.NotBeforeDuration.Duration),
	}, nil
}

func (in *VaultCertificateRole) GetSubject() (*pki.SubjectSettings, error) {
	return &pki.SubjectSettings{
		Organization:       in.Spec.Subject.Organization,
		OrganizationalUnit: in.Spec.Subject.OrganizationalUnit,
		Country:            in.Spec.Subject.Country,
		Locality:           in.Spec.Subject.Locality,
		Province:           in.Spec.Subject.Province,
		StreetAddress:      in.Spec.Subject.StreetAddress,
		PostalCode:         in.Spec.Subject.PostalCode,
	}, nil
}
