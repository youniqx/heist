package common

import (
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InternalKvEngineMountPath = "managed/_heist_internal"

// InternalKvEngine is the KV Secret Engine used by Heist to persist internal data.
var InternalKvEngine = &kvengine.KvEngine{
	Path: InternalKvEngineMountPath,
	Config: &kvengine.Config{
		MaxVersions:        0,
		CasRequired:        true,
		DeleteVersionAfter: "",
	},
}

const (
	// CAPrivateKeyField is the field name in an internal secret containing a private key.
	CAPrivateKeyField = "private_key"

	// CAPrivateKeyTypeField is the field name in an internal secret containing the private key type.
	CAPrivateKeyTypeField = "private_key_type"

	// CACertificateField is the field name in an internal secret containing the certificate.
	CACertificateField = "certificate"

	// CACertificateChainField is the field name in an internal secret containing the certificate chain.
	CACertificateChainField = "certificate_chain"

	// CACertificateFullChainField is the field name in an internal secret containing the full cert chain (including root cert).
	CACertificateFullChainField = "full_certificate_chain"

	// CAIssuerField is the field name in an internal secret containing the issuing certificate.
	CAIssuerField = "issuer"

	// CASerialNumberField is the field name in an internal secret containing the certificate serial number.
	CASerialNumberField = "serial_number"
)

func GetCAInfoSecretPath(ca *heistv1alpha1.VaultCertificateAuthority) string {
	return fmt.Sprintf("%s/pki/ca/public/%s", ca.Namespace, ca.Name)
}

func GetCAPrivateKeySecretPath(ca *heistv1alpha1.VaultCertificateAuthority) string {
	return fmt.Sprintf("%s/pki/ca/private/%s", ca.Namespace, ca.Name)
}

func GetAnnotationValue(object client.Object, annotations ...string) (string, bool) {
	objAnnotations := object.GetAnnotations()
	if objAnnotations == nil {
		return "", false
	}
	for _, annotation := range annotations {
		if value, ok := objAnnotations[annotation]; ok {
			return value, true
		}
	}
	return "", false
}
