package common

import (
	"fmt"

	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
)

func GetPolicyNameForSecret(secret *heistv1alpha1.VaultKVSecret) string {
	return fmt.Sprintf("managed.kv.%s.%s", secret.Namespace, secret.Name)
}

func GetPolicyNameForCertificateIssuing(cert *heistv1alpha1.VaultCertificateRole) string {
	return fmt.Sprintf("managed.pki.cert.issue.%s.%s", cert.Namespace, cert.Name)
}

func GetPolicyNameForCertificateSignCSR(cert *heistv1alpha1.VaultCertificateRole) string {
	return fmt.Sprintf("managed.pki.cert.sign-csr.%s.%s", cert.Namespace, cert.Name)
}

func GetPolicyNameForCertificateSignVerbatim(cert *heistv1alpha1.VaultCertificateRole) string {
	return fmt.Sprintf("managed.pki.cert.sign-verbatim.%s.%s", cert.Namespace, cert.Name)
}

func GetPolicyNameForCertificateAuthorityPrivateInfo(ca *heistv1alpha1.VaultCertificateAuthority) string {
	return fmt.Sprintf("managed.pki.ca.private.%s.%s", ca.Namespace, ca.Name)
}

func GetPolicyNameForCertificateAuthorityPublicInfo(ca *heistv1alpha1.VaultCertificateAuthority) string {
	return fmt.Sprintf("managed.pki.ca.public.%s.%s", ca.Namespace, ca.Name)
}

func GetPolicyNameForTransitKeyEncrypt(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.encrypt.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyDecrypt(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.decrypt.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyRewrap(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.rewrap.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyDatakey(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.datakey.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyHmac(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.hmac.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeySign(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.sign.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyVerify(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.verify.%s.%s", key.Namespace, key.Name)
}

func GetPolicyNameForTransitKeyRead(key *heistv1alpha1.VaultTransitKey) string {
	return fmt.Sprintf("managed.transit.key.read.%s.%s", key.Namespace, key.Name)
}
