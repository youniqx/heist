/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vaultcertificaterole

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault/pki"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VaultCertificateRole Controller", func() {
	When("Trying to certificate a certificate", func() {
		var (
			issuer *heistv1alpha1.VaultCertificateAuthority
			cert   *heistv1alpha1.VaultCertificateRole
		)

		BeforeEach(func() {
			issuer = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						CommonName: "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
					},
				},
			}
			Test.K8sEnv.Create(issuer)

			cert = &heistv1alpha1.VaultCertificateRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("my-cert-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateRoleSpec{
					Issuer: issuer.Name,
					Settings: heistv1alpha1.VaultCertificateRoleSettings{
						KeyType: pki.KeyTypeRSA,
						KeyBits: pki.KeyBitsRSA2048,
					},
				},
			}
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should not exist before creating it", func() {
			Test.VaultEnv.CertificateRole(issuer, cert).Should(BeNil())
		})

		It("Should be able to be provisioned", func() {
			Test.K8sEnv.Create(cert)
			Test.VaultEnv.CertificateRole(issuer, cert).ShouldNot(BeNil())
		})
	})
})
