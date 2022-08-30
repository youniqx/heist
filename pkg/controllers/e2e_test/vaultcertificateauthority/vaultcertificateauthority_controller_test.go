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

package vaultcertificateauthority

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	"github.com/youniqx/heist/pkg/controllers/e2e_test"
	. "github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault/core"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	"github.com/youniqx/heist/pkg/vault/pki"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VaultCertificateAuthority Controller", func() {
	When("creating an internal root CA", func() {
		var ca *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			ca = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
				},
			}
			Test.K8sEnv.Create(ca)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be provisioned", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))

			Test.VaultEnv.CA(ca).ShouldNot(BeNil())
			Test.VaultEnv.CA(ca).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", ca.Namespace, ca.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("certificate_chain", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("full_certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("issuer"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("serial_number"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key_type", 0))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).ShouldNot(BeNil())
		})

		It("Should delete the associated info secrets after the CA has been deleted", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(ca)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).Should(BeNil())
		})
	})

	When("importing an internal root CA", func() {
		var ca *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			ca = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.RootCertificateCipherText,
						PrivateKey:  Test.RootPrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(ca)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be provisioned", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(ca).ShouldNot(BeNil())
			Test.VaultEnv.CA(ca).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", ca.Namespace, ca.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("certificate_chain", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("full_certificate_chain", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("issuer", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("serial_number", "39:af:8c:ff:af:94:27:5f:49:7f:91:99:cc:ad:2e:cc:a3:bf:15:d7"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key_type", 0))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).ShouldNot(BeNil())
		})

		It("Should delete the associated info secrets after the CA has been deleted", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(ca)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).Should(BeNil())
		})
	})

	When("creating an exported root CA", func() {
		var ca *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			ca = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                true,
					},
				},
			}
			Test.K8sEnv.Create(ca)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be created with the private key persisted in the internal kv engine", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(ca).ShouldNot(BeNil())
			Test.VaultEnv.CA(ca).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", ca.Namespace, ca.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("certificate_chain", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("full_certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("issuer"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("serial_number"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretField("private_key"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretField("private_key_type"))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).ShouldNot(BeNil())
		})

		It("Should delete the associated infos secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(ca)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).Should(BeNil())
		})
	})

	When("importing an exported root CA", func() {
		var ca *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			ca = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                true,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.RootCertificateCipherText,
						PrivateKey:  Test.RootPrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(ca)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be created with the private key persisted in the internal kv engine", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(ca).ShouldNot(BeNil())
			Test.VaultEnv.CA(ca).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", ca.Namespace, ca.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("certificate_chain", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("full_certificate_chain", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("issuer", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("serial_number", "39:af:8c:ff:af:94:27:5f:49:7f:91:99:cc:ad:2e:cc:a3:bf:15:d7"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldWithValue("private_key", e2e_test.RootPrivateKey))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldWithValue("private_key_type", "rsa"))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).ShouldNot(BeNil())
		})

		It("Should delete the associated infos secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(ca).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(ca))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(ca))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(ca)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", ca.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", ca.Name))).Should(BeNil())
		})
	})

	When("creating an internal intermediate CA", func() {
		var intermediateCA *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			rootCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
				},
			}
			Test.K8sEnv.Create(rootCA)

			intermediateCA = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-intermediate-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
				},
			}
			Test.K8sEnv.Create(intermediateCA)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be provisioned", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))

			Test.VaultEnv.CA(intermediateCA).ShouldNot(BeNil())
			Test.VaultEnv.CA(intermediateCA).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", intermediateCA.Namespace, intermediateCA.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("full_certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("issuer"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("serial_number"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key_type", 0))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
		})

		It("Should delete associated info secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(intermediateCA)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).Should(BeNil())
		})
	})

	When("importing an internal intermediate CA", func() {
		var intermediateCA *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			rootCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.RootCertificateCipherText,
						PrivateKey:  Test.RootPrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(rootCA)

			intermediateCA = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-intermediate-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.IntermediateCertificateCipherText,
						PrivateKey:  Test.IntermediatePrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(intermediateCA)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be provisioned", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(intermediateCA).ShouldNot(BeNil())
			Test.VaultEnv.CA(intermediateCA).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", intermediateCA.Namespace, intermediateCA.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate", e2e_test.IntermediateCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate_chain", e2e_test.IntermediateCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("full_certificate_chain", e2e_test.IntermediateCertificate+"\n"+e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("issuer", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("serial_number", "c2:fd:c4:66:b3:c0:e0:61"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key", 0))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldFieldWithLength("private_key_type", 0))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
		})

		It("Should delete associated info secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(intermediateCA)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).Should(BeNil())
		})
	})

	When("creating an exported intermediate CA", func() {
		var intermediateCA *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			rootCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
				},
			}
			Test.K8sEnv.Create(rootCA)

			intermediateCA = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-intermediate-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                true,
					},
				},
			}
			Test.K8sEnv.Create(intermediateCA)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be created with the public and private info persisted in the internal kv engine", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(intermediateCA).ShouldNot(BeNil())
			Test.VaultEnv.CA(intermediateCA).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", intermediateCA.Namespace, intermediateCA.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("full_certificate_chain"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("issuer"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretField("serial_number"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretField("private_key"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretField("private_key_type"))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
		})

		It("Should delete associated info secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(intermediateCA)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).Should(BeNil())
		})
	})

	When("importing an exported intermediate CA", func() {
		var intermediateCA *heistv1alpha1.VaultCertificateAuthority

		BeforeEach(func() {
			rootCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-root-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                false,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.RootCertificateCipherText,
						PrivateKey:  Test.RootPrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(rootCA)

			intermediateCA = &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						Organization:       []string{},
						OrganizationalUnit: []string{},
						Country:            []string{},
						Locality:           []string{},
						Province:           []string{},
						StreetAddress:      []string{},
						PostalCode:         []string{},
						CommonName:         "my-intermediate-ca",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						SubjectAlternativeNames: []string{"test.com"},
						IPSans:                  []string{},
						URISans:                 []string{},
						OtherSans:               []string{},
						TTL:                     metav1.Duration{},
						KeyType:                 pki.KeyTypeRSA,
						KeyBits:                 pki.KeyBitsRSA2048,
						ExcludeCNFromSans:       true,
						PermittedDNSDomains:     []string{},
						Exported:                true,
					},
					Import: &heistv1alpha1.VaultCertificateAuthorityImport{
						Certificate: Test.IntermediateCertificateCipherText,
						PrivateKey:  Test.IntermediatePrivateKeyCipherText,
					},
				},
			}
			Test.K8sEnv.Create(intermediateCA)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("Should be created with the public and private info persisted in the internal kv engine", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			Test.VaultEnv.CA(intermediateCA).ShouldNot(BeNil())
			Test.VaultEnv.CA(intermediateCA).Should(HavePath(fmt.Sprintf("managed/pki/%s/%s", intermediateCA.Namespace, intermediateCA.Name)))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate", e2e_test.IntermediateCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("certificate_chain", e2e_test.IntermediateCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("full_certificate_chain", e2e_test.IntermediateCertificate+"\n"+e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("issuer", e2e_test.RootCertificate))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(HaveKvSecretFieldWithValue("serial_number", "c2:fd:c4:66:b3:c0:e0:61"))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldWithValue("private_key", e2e_test.IntermediatePrivateKey))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(HaveKvSecretFieldWithValue("private_key_type", "rsa"))

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).ShouldNot(BeNil())
		})

		It("Should delete associated info secrets and policies after the CA has been deleted", func() {
			Test.K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			publicInfoSecretPath := core.SecretPath(common.GetCAInfoSecretPath(intermediateCA))
			privateInfoSecretPath := core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA))
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).ShouldNot(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).ShouldNot(BeNil())
			Test.K8sEnv.DeleteIfPresent(intermediateCA)
			Test.VaultEnv.KvSecret(common.InternalKvEngine, publicInfoSecretPath).Should(BeNil())
			Test.VaultEnv.KvSecret(common.InternalKvEngine, privateInfoSecretPath).Should(BeNil())

			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.private.default.%s", intermediateCA.Name))).Should(BeNil())
			Test.VaultEnv.Policy(core.PolicyName(fmt.Sprintf("managed.pki.ca.public.default.%s", intermediateCA.Name))).Should(BeNil())
		})
	})
})
