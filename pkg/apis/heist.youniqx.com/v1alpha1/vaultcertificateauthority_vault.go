package v1alpha1

import "fmt"

func (in *VaultCertificateAuthority) GetMountPath() (string, error) {
	return fmt.Sprintf("managed/pki/%s/%s", in.Namespace, in.Name), nil
}
