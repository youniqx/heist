package pki

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/mount"
)

type EngineAPI interface {
	UpdatePKIEngine(engine EngineEntity) error
	ReadPKIEngine(engine core.MountPathEntity) (*Engine, error)
	ListCerts(engine core.MountPathEntity) ([]string, error)
	ReadCACertificatePEM(ca core.MountPathEntity) (string, error)
}

type CAInfo struct {
	Path                        string
	SerialNumber                string
	PrivateKey                  string
	PrivateKeyType              KeyType
	IssuingCertificateAuthority string
	CertificateChain            string
	Certificate                 string
}

type TidySettings struct {
	TidyCertStore    bool
	TidyRevokedCerts bool
	SafetyBuffer     *core.VaultTTL
}

type SignCsr struct {
	CSR              string        `json:"csr"`
	CommonName       string        `json:"common_name"`
	AlternativeNames []string      `json:"alternative_names"`
	OtherSans        []string      `json:"other_sans"`
	IPSans           []string      `json:"ip_sans"`
	URISans          []string      `json:"uri_sans"`
	TTL              time.Duration `json:"ttl"`
}

type API interface {
	EngineAPI
	IsPKIEngineInitialized(ca core.MountPathEntity) (bool, error)
	CreateRootCA(mode Mode, ca CAEntity) (*CAInfo, error)
	UpdateRootCA(ca CAEntity) error
	CreateIntermediateCA(mode Mode, issuer core.MountPathEntity, ca CAEntity) (*CAInfo, error)
	UpdateIntermediateCA(issuer core.MountPathEntity, ca CAEntity) error
	ReadCA(ca core.MountPathEntity) (*CA, error)
	UpdateCertificateRole(ca core.MountPathEntity, role CertificateRoleEntity) error
	ReadCertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) (*CertificateRole, error)
	DeleteCertificateRole(ca core.MountPathEntity, role core.RoleNameEntity) error
	SignCertificateSigningRequest(ca core.MountPathEntity, role core.RoleNameEntity, request *SignCsr) (*Certificate, error)
	IssueCertificate(ca core.MountPathEntity, role core.RoleNameEntity, options *IssueCertOptions) (*Certificate, error)
	RevokeCertificate(ca core.MountPathEntity, serial SerialNumberEntity) error
	Tidy(ca core.MountPathEntity, settings *TidySettings) error
	RotateCRLs(ca core.MountPathEntity) error
}

type KeyType string

const (
	// KeyTypeRSA defines that the key type is rsa.
	KeyTypeRSA KeyType = "rsa"
	// KeyTypeEC defines that the key type is ec.
	KeyTypeEC KeyType = "ec"
	// KeyTypeAny defines that the key type can be any vault supported type.
	KeyTypeAny KeyType = "any"
)

type KeyBits int

const (
	// KeyBitsEC224 defines that the EC key has 224 bits.
	KeyBitsEC224 = 224
	// KeyBitsEC256 defines that the EC key has 256 bits.
	KeyBitsEC256 = 256
	// KeyBitsEC384 defines that the EC key has 384 bits.
	KeyBitsEC384 = 384
	// KeyBitsEC521 defines that the EC key has 521 bits.
	KeyBitsEC521 = 521
	// KeyBitsRSA2048 defines that the RSA key has 2048 bits.
	KeyBitsRSA2048 = 2048
	// KeyBitsRSA3072 defines that the RSA key has 3072 bits.
	KeyBitsRSA3072 = 3072
	// KeyBitsRSA4096 defines that the RSA key has 4096 bits.
	KeyBitsRSA4096 = 4096
)

type KeyUsage string

const (
	// KeyUsageDigitalSignature configures the key usage DigitalSignature.
	KeyUsageDigitalSignature KeyUsage = "DigitalSignature"
	// KeyUsageKeyAgreement configures the key usage KeyAgreement.
	KeyUsageKeyAgreement KeyUsage = "KeyAgreement"
	// KeyUsageKeyEncipherment configures the key usage KeyEncipherment.
	KeyUsageKeyEncipherment KeyUsage = "KeyEncipherment"
	// KeyUsageContentCommitment configures the key usage ContentCommitment.
	KeyUsageContentCommitment KeyUsage = "ContentCommitment"
	// KeyUsageDataEncipherment configures the key usage DataEncipherment.
	KeyUsageDataEncipherment KeyUsage = "DataEncipherment"
	// KeyUsageCertSign configures the key usage CertSign.
	KeyUsageCertSign KeyUsage = "CertSign"
	// KeyUsageCRLSign configures the key usage CRLSign.
	KeyUsageCRLSign KeyUsage = "CRLSign"
	// KeyUsageEncipherOnly configures the key usage EncipherOnly.
	KeyUsageEncipherOnly KeyUsage = "EncipherOnly"
	// KeyUsageDecipherOnly configures the key usage DecipherOnly.
	KeyUsageDecipherOnly KeyUsage = "DecipherOnly"
)

type ExtendedKeyUsage string

const (
	// ExtendedKeyUsageAny configures the extended key usage Any.
	ExtendedKeyUsageAny ExtendedKeyUsage = "Any"
	// ExtendedKeyUsageServerAuth configures the extended key usage ServerAuth.
	ExtendedKeyUsageServerAuth ExtendedKeyUsage = "ServerAuth"
	// ExtendedKeyUsageClientAuth configures the extended key usage ClientAuth.
	ExtendedKeyUsageClientAuth ExtendedKeyUsage = "ClientAuth"
	// ExtendedKeyUsageCodeSigning configures the extended key usage CodeSigning.
	ExtendedKeyUsageCodeSigning ExtendedKeyUsage = "CodeSigning"
	// ExtendedKeyUsageEmailProtection configures the extended key usage EmailProtection.
	ExtendedKeyUsageEmailProtection ExtendedKeyUsage = "EmailProtection"
	// ExtendedKeyUsageIPSECEndSystem configures the extended key usage IPSECEndSystem.
	ExtendedKeyUsageIPSECEndSystem ExtendedKeyUsage = "IPSECEndSystem"
	// ExtendedKeyUsageIPSECTunnel configures the extended key usage IPSECTunnel.
	ExtendedKeyUsageIPSECTunnel ExtendedKeyUsage = "IPSECTunnel"
	// ExtendedKeyUsageIPSECUser configures the extended key usage IPSECUser.
	ExtendedKeyUsageIPSECUser ExtendedKeyUsage = "IPSECUser"
	// ExtendedKeyUsageTimeStamping configures the extended key usage TimeStamping.
	ExtendedKeyUsageTimeStamping ExtendedKeyUsage = "TimeStamping"
	// ExtendedKeyUsageOCSPSigning configures the extended key usage OCSPSigning.
	ExtendedKeyUsageOCSPSigning ExtendedKeyUsage = "OCSPSigning"
	// ExtendedKeyUsageMicrosoftServerGatedCrypto configures the extended key usage MicrosoftServerGatedCrypto.
	ExtendedKeyUsageMicrosoftServerGatedCrypto ExtendedKeyUsage = "MicrosoftServerGatedCrypto"
	// ExtendedKeyUsageNetscapeServerGatedCrypto configures the extended key usage NetscapeServerGatedCrypto.
	ExtendedKeyUsageNetscapeServerGatedCrypto ExtendedKeyUsage = "NetscapeServerGatedCrypto"
	// ExtendedKeyUsageMicrosoftCommercialCodeSigning configures the extended key usage MicrosoftCommercialCodeSigning.
	ExtendedKeyUsageMicrosoftCommercialCodeSigning ExtendedKeyUsage = "MicrosoftCommercialCodeSigning"
	// ExtendedKeyUsageMicrosoftKernelCodeSigning configures the extended key usage MicrosoftKernelCodeSigning.
	ExtendedKeyUsageMicrosoftKernelCodeSigning ExtendedKeyUsage = "MicrosoftKernelCodeSigning"
)

type StringArray []string

func (s *StringArray) UnmarshalJSON(b []byte) error {
	var sn []string
	if err := json.Unmarshal(b, &sn); err != nil {
		return core.ErrAPIError.WithDetails("couldn't unmarshal field").WithCause(err)
	}

	*s = sn

	return nil
}

func (s StringArray) MarshalJSON() ([]byte, error) {
	j, err := json.Marshal(strings.Join(s, ","))
	if err != nil {
		return nil, core.ErrAPIError.WithDetails("couldn't marshal field").WithCause(err)
	}

	return j, nil
}

type SubjectSettings struct {
	// Organization specifies the O values in the subject field of the resulting certificate.
	Organization StringArray `json:"organization"`
	// OrganizationalUnit specifies the OU (OrganizationalUnit) values in the subject field
	// of the resulting certificate.
	OrganizationalUnit StringArray `json:"ou"`
	// Country specifies the C values in the subject field of the resulting certificate.
	Country StringArray `json:"country"`
	// Locality specifies the L values in the subject field of the resulting certificate.
	Locality StringArray `json:"locality"`
	// Province specifies the ST values in the subject field of the resulting certificate.
	Province StringArray `json:"province"`
	// StreetAddress specifies the Street Address values in the subject field of the resulting
	// certificate.
	StreetAddress StringArray `json:"street_address"`
	// PostalCode specifies the Postal Code values in the subject field of the resulting
	// certificate.
	PostalCode StringArray `json:"postal_code"`
}

type Subject struct {
	*SubjectSettings
	// CommonName specifies the requested CN for the certificate.
	CommonName string `json:"common_name"`
	// SerialNumber specifies the Serial Number, if any. Otherwise Vault will generate a random serial
	// for you. If you want more than one, specify alternative names in the alt_names map using
	// OID 2.5.4.5.
	SerialNumber string `json:"serial_number"`
}

type CASettings struct {
	// SubjectAlternativeNames Specifies the requested Subject Alternative Names, in a
	// comma-delimited list. These can be host names or email addresses; they will be
	// parsed into their respective fields.
	SubjectAlternativeNames StringArray `json:"alt_names"`
	// IPSans specifies the requested IP Subject Alternative Names, in a comma-delimited
	// list.
	IPSans StringArray `json:"ip_sans"`
	// URISans specifies the requested URI Subject Alternative Names, in a comma-delimited
	// list.
	URISans StringArray `json:"uri_sans"`
	// OtherSans specifies custom OID/UTF8-string SANs. These must match values specified on
	// the role in allowed_other_sans (see role creation for allowed_other_sans globbing rules).
	// The format is the same as OpenSSL: <oid>;<type>:<value> where the only current valid
	// type is UTF8.
	OtherSans StringArray `json:"other_sans"`
	// TTL specifies the requested Time To Live (after which the certificate will be expired).
	// This cannot be larger than the engine's max (or, if not set, the system max).
	TTL *core.VaultTTL `json:"ttl"`
	// KeyType specifies the desired key type; must be rsa or ec.
	KeyType KeyType `json:"key_type"`
	// KeyBits specifies the number of bits to use. This must be changed to a valid value if the
	// KeyType is ec, e.g., 224, 256, 384 or 521.
	KeyBits KeyBits `json:"key_bits"`
	// ExcludeCNFromSans specifies that the given common_name will not be included in DNS or Email
	// SubjectAlternativeNames (as appropriate). Useful if the CN is not a hostname or email address,
	// but is instead some human-readable identifier.
	ExcludeCNFromSans bool `json:"exclude_cn_from_sans"`
	// PermittedDNSDomains specifies DNS domains for which certificates are allowed to be issued or
	// signed by this CA certificate. Note that subdomains are allowed, as per RFC.
	PermittedDNSDomains StringArray `json:"permitted_dns_domains"`
}

type CA struct {
	Path         string
	Settings     *CASettings
	Subject      *Subject
	PluginName   string
	Config       *mount.TuneConfig
	ImportedCert *ImportedCert
}

func (c *CA) GetPKIEngineConfig() (*mount.TuneConfig, error) {
	return c.Config, nil
}

func (c *CA) GetMountPath() (string, error) {
	return c.Path, nil
}

func (c *CA) GetPluginName() (string, error) {
	return c.PluginName, nil
}

func (c *CA) GetSettings() (*CASettings, error) {
	return c.Settings, nil
}

func (c *CA) GetSubject() (*Subject, error) {
	return c.Subject, nil
}

func (c *CA) GetImportedCert() (*ImportedCert, error) {
	return c.ImportedCert, nil
}

type Mode string

const (
	// ModeInternal configures the CA to use internal mode. With internal CAs it is not possible
	// to fetch the private key in plain text. It can only be used to perform crypto operations
	// inside vault.
	ModeInternal Mode = "internal"
	// ModeExported configures the CA to use exported mode. With exported CAs it is possible
	// to fetch the private key in plain text to use it outside of vault.
	ModeExported Mode = "exported"
)

type ImportedCert struct {
	PrivateKey  string `json:"private_key"`
	Certificate string `json:"certificate"`
}

type CAEntity interface {
	EngineEntity
	GetSettings() (*CASettings, error)
	GetSubject() (*Subject, error)
	GetImportedCert() (*ImportedCert, error)
}

type RoleSettings struct {
	TTL                           *core.VaultTTL     `json:"ttl"`
	MaxTTL                        *core.VaultTTL     `json:"max_ttl"`
	AllowLocalhost                bool               `json:"allow_localhost"`
	AllowedDomains                []string           `json:"allowed_domains"`
	AllowedDomainsTemplate        bool               `json:"allowed_domains_template"`
	AllowBareDomains              bool               `json:"allow_bare_domains"`
	AllowSubdomains               bool               `json:"allow_subdomains"`
	AllowGlobDomains              bool               `json:"allow_glob_domains"`
	AllowAnyName                  bool               `json:"allow_any_name"`
	EnforceHostNames              bool               `json:"enforce_hostnames"`
	AllowIPSans                   bool               `json:"allow_ip_sans"`
	AllowedURISans                []string           `json:"allowed_uri_sans"`
	AllowedOtherSans              []string           `json:"allowed_other_sans"`
	ServerFlag                    bool               `json:"server_flag"`
	ClientFlag                    bool               `json:"client_flag"`
	CodeSigningFlag               bool               `json:"code_signing_flag"`
	EmailProtectionFlag           bool               `json:"email_protection_flag"`
	KeyType                       KeyType            `json:"key_type"`
	KeyBits                       KeyBits            `json:"key_bits"`
	KeyUsage                      []KeyUsage         `json:"key_usage"`
	ExtendedKeyUsage              []ExtendedKeyUsage `json:"ext_key_usage"`
	ExtendedKeyUsageOids          []string           `json:"ext_key_usage_oids"`
	UseCSRCommonName              bool               `json:"use_csr_common_name"`
	UseCSRSans                    bool               `json:"use_csr_sans"`
	GenerateLease                 bool               `json:"generate_lease"`
	NoStore                       bool               `json:"no_store"`
	RequireCommonName             bool               `json:"require_cn"`
	PolicyIdentifiers             []string           `json:"policy_identifiers"`
	BasicConstraintsValidForNonCA bool               `json:"basic_constraints_valid_for_non_ca"`
	NotBeforeDuration             *core.VaultTTL     `json:"not_before_duration"`
}

type CertificateRoleEntity interface {
	core.RoleNameEntity
	GetSettings() (*RoleSettings, error)
	GetSubject() (*SubjectSettings, error)
}

type CertificateRole struct {
	Name     string
	Settings *RoleSettings
	Subject  *SubjectSettings
}

func (c *CertificateRole) GetRoleName() (string, error) {
	return c.Name, nil
}

func (c *CertificateRole) GetSettings() (*RoleSettings, error) {
	return c.Settings, nil
}

func (c *CertificateRole) GetSubject() (*SubjectSettings, error) {
	return c.Subject, nil
}

type SerialNumberEntity interface {
	GetSerialNumber() (string, error)
}

type SerialNumber string

func (s SerialNumber) GetSerialNumber() (string, error) {
	return string(s), nil
}

type CertificateEntity interface {
	SerialNumberEntity
	GetIssuingCA() (string, error)
	GetCAChain() ([]string, error)
	GetPrivateKey() (string, error)
	GetPrivateKeyType() (KeyType, error)
}

type Certificate struct {
	Certificate    string   `json:"certificate"`
	IssuingCA      string   `json:"issuing_ca"`
	CAChain        []string `json:"ca_chain"`
	PrivateKey     string   `json:"private_key"`
	PrivateKeyType KeyType  `json:"private_key_type"`
	SerialNumber   string   `json:"serial_number"`
}

func (c *Certificate) GetSerialNumber() (string, error) {
	return c.SerialNumber, nil
}

func (c *Certificate) GetIssuingCA() (string, error) {
	return c.IssuingCA, nil
}

func (c *Certificate) GetCAChain() ([]string, error) {
	return c.CAChain, nil
}

func (c *Certificate) GetPrivateKey() (string, error) {
	return c.PrivateKey, nil
}

func (c *Certificate) GetPrivateKeyType() (KeyType, error) {
	return c.PrivateKeyType, nil
}

type pkiAPI struct {
	Core  core.API
	Mount mount.API
}

func NewAPI(coreAPI core.API, mountAPI mount.API) API {
	return &pkiAPI{
		Core:  coreAPI,
		Mount: mountAPI,
	}
}
