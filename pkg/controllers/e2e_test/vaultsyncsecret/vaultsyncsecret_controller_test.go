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

package vaultsyncsecret

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers/common"
	. "github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/pki"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VaultSyncSecret Controller", func() {
	When("trying to sync value from a certificate authority", func() {
		var (
			rootCA         *heistv1alpha1.VaultCertificateAuthority
			intermediateCA *heistv1alpha1.VaultCertificateAuthority
			cert           *heistv1alpha1.VaultCertificateRole
		)

		BeforeEach(func() {
			rootCA = &heistv1alpha1.VaultCertificateAuthority{
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

			cert = &heistv1alpha1.VaultCertificateRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("my-cert-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateRoleSpec{
					Issuer: intermediateCA.Name,
					Settings: heistv1alpha1.VaultCertificateRoleSettings{
						TTL:              metav1.Duration{Duration: time.Minute * 10},
						AllowBareDomains: true,
						AllowSubdomains:  true,
						AllowGlobDomains: true,
						AllowedDomains:   []string{"example.com"},
						KeyType:          pki.KeyTypeRSA,
						KeyBits:          pki.KeyBitsRSA2048,
					},
				},
			}
			Test.K8sEnv.Create(cert)
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("should sync the full cert chain of an intermediate ca", func() {
			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"full_chain": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())

			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["full_chain"]).NotTo(BeEmpty())

			kvSecret, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(kvSecret).NotTo(BeNil())
			Expect(kvSecret.Fields["full_certificate_chain"]).To(Equal(string(secret.Data["full_chain"])))
		})

		It("should sync the cert chain of an intermediate ca", func() {
			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"cert_chain": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertChain,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())

			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["cert_chain"]).NotTo(BeEmpty())

			kvSecret, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(kvSecret).NotTo(BeNil())
			Expect(kvSecret.Fields["certificate_chain"]).To(Equal(string(secret.Data["cert_chain"])))
		})

		It("should sync the certificate of an intermediate ca", func() {
			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"cert": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertificate,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())

			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["cert"]).NotTo(BeEmpty())

			kvSecret, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(kvSecret).NotTo(BeNil())
			Expect(kvSecret.Fields["certificate"]).To(Equal(string(secret.Data["cert"])))
		})

		It("should sync the private key of an intermediate ca", func() {
			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"key": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())

			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["key"]).NotTo(BeEmpty())

			kvSecret, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(kvSecret).NotTo(BeNil())
			Expect(kvSecret.Fields["private_key"]).To(Equal(string(secret.Data["key"])))
		})

		It("should refuse to sync if the secret already exists", func() {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Data: map[string][]byte{
					"some_key": []byte("some_value"),
				},
			}
			Test.K8sEnv.Create(secret)

			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      secret.Name,
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"key": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionFalse,
				heistv1alpha1.Conditions.Reasons.ErrorConfig,
				"Secret already exists or is manged by someone else",
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())
			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["some_key"]).To(Equal([]byte("some_value")))
		})

		It("should refuse to sync if the target namespace is not on the allow list", func() {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "disallowed",
				},
				Data: map[string][]byte{
					"some_key": []byte("some_value"),
				},
			}

			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      secret.Name,
						Namespace: secret.Name,
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"key": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionFalse,
				heistv1alpha1.Conditions.Reasons.ErrorConfig,
				fmt.Sprintf("Namespace %s of secret is not allowed", secret.Name),
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).NotTo(Succeed())
		})

		It("should sync values from an issued certificate", func() {
			sync := &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
						Type:      v1.SecretTypeTLS,
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{
						{
							CertificateRole: cert.Name,
							CommonName:      "example.com",
							DNSSans: []string{
								"*.example.com",
								"example.com",
							},
							ExcludeCNFromSans: false,
						},
					},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"full_chain": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  cert.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain,
							},
						},
						"cert_chain": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  cert.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertChain,
							},
						},
						"tls.crt": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  cert.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertificate,
							},
						},
						"tls.key": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  cert.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())

			Expect(secret.Data).To(HaveLen(4))
			Expect(secret.Type).To(Equal(v1.SecretTypeTLS))
			Expect(secret.Data["full_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["cert_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["tls.crt"]).NotTo(BeEmpty())
			Expect(secret.Data["tls.key"]).NotTo(BeEmpty())

			Expect(secret.Data["full_chain"]).NotTo(Equal(secret.Data["cert_chain"]))
			Expect(secret.Data["cert_chain"]).NotTo(Equal(secret.Data["tls.crt"]))
		})
	})

	When("trying to sync value from a kv secret", func() {
		var (
			engine        *heistv1alpha1.VaultKVSecretEngine
			vaultKVSecret *heistv1alpha1.VaultKVSecret
			sync          *heistv1alpha1.VaultSyncSecret
			secret        *v1.Secret
		)

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("engine-%d", time.Now().Unix()),
					Namespace: "default",
				},
			}
			Test.K8sEnv.Create(engine)

			vaultKVSecret = &heistv1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("secret-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some_field": {
							CipherText: heistv1alpha1.EncryptedValue(Test.DefaultCipherText),
						},
					},
				},
			}
			Test.K8sEnv.Create(vaultKVSecret)

			Test.K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))
			Test.K8sEnv.Object(vaultKVSecret).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been provisioned",
			))

			secret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("secret-%d", time.Now().Unix()),
					Namespace: "default",
				},
			}
			sync = &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name: secret.Name,
					},
					CertificateTemplates: nil,
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"some_key": {
							KVSecret: &heistv1alpha1.VaultSyncKVSecretSource{
								Name:  vaultKVSecret.Name,
								Field: "some_field",
							},
						},
					},
				},
			}
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("should be able to sync the secret", func() {
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())
			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.Data["some_key"]).To(Equal([]byte("ASDF ASDF")))
		})

		It("should sync secret with labels and annotations", func() {
			testResource := sync.DeepCopy()

			testResource.Spec.Target.AdditionalAnnotations = map[string]string{
				"youniqx.com/test-annotation": "true",
			}
			testResource.Spec.Target.AdditionalLabels = map[string]string{
				"youniqx.com/test-label": "true",
			}

			Test.K8sEnv.Create(testResource)

			Test.K8sEnv.Object(testResource).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())
			Expect(secret.Data).To(HaveLen(1))
			Expect(secret.ObjectMeta.Annotations["youniqx.com/test-annotation"]).To(Equal("true"))
			Expect(secret.ObjectMeta.Labels["youniqx.com/test-label"]).To(Equal("true"))
			Expect(secret.Data["some_key"]).To(Equal([]byte("ASDF ASDF")))
		})
	})

	When("managing a sync secret resource", func() {
		var (
			rootCA         *heistv1alpha1.VaultCertificateAuthority
			intermediateCA *heistv1alpha1.VaultCertificateAuthority
			cert           *heistv1alpha1.VaultCertificateRole
			engine         *heistv1alpha1.VaultKVSecretEngine
			vaultKVSecret  *heistv1alpha1.VaultKVSecret
			sync           *heistv1alpha1.VaultSyncSecret
			secret         *v1.Secret
		)

		BeforeEach(func() {
			rootCA = &heistv1alpha1.VaultCertificateAuthority{
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

			engine = &heistv1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("engine-%d", time.Now().Unix()),
					Namespace: "default",
				},
			}
			Test.K8sEnv.Create(engine)

			vaultKVSecret = &heistv1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("secret-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some_field": {
							CipherText: heistv1alpha1.EncryptedValue(Test.DefaultCipherText),
						},
					},
				},
			}
			Test.K8sEnv.Create(vaultKVSecret)

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

			cert = &heistv1alpha1.VaultCertificateRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("my-cert-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateRoleSpec{
					Issuer: intermediateCA.Name,
					Settings: heistv1alpha1.VaultCertificateRoleSettings{
						TTL:              metav1.Duration{Duration: time.Minute * 10},
						AllowBareDomains: true,
						AllowSubdomains:  true,
						AllowGlobDomains: true,
						AllowedDomains:   []string{"example.com"},
						KeyType:          pki.KeyTypeRSA,
						KeyBits:          pki.KeyBitsRSA2048,
					},
				},
			}
			Test.K8sEnv.Create(cert)

			sync = &heistv1alpha1.VaultSyncSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultSyncSecretSpec{
					Target: heistv1alpha1.VaultSyncSecretTarget{
						Name:      fmt.Sprintf("sync-%d", time.Now().Unix()),
						Namespace: "default",
					},
					CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{
						{
							Alias:           "template_0",
							CertificateRole: cert.Name,
							CommonName:      "example.com",
							DNSSans: []string{
								"*.example.com",
								"example.com",
							},
							ExcludeCNFromSans: false,
						},
					},
					Data: map[string]heistv1alpha1.VaultSyncSecretSource{
						"cert_chain": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertChain,
							},
						},
						"full_chain": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain,
							},
						},
						"cert": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertificate,
							},
						},
						"key": {
							CertificateAuthority: &heistv1alpha1.VaultSyncCertificateAuthoritySource{
								Name:  intermediateCA.Name,
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
						"some_key": {
							KVSecret: &heistv1alpha1.VaultSyncKVSecretSource{
								Name:  vaultKVSecret.Name,
								Field: "some_field",
							},
						},
						"issued_full_chain": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  "template_0",
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeFullCertChain,
							},
						},
						"issued_cert_chain": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  "template_0",
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertChain,
							},
						},
						"tls.crt": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  "template_0",
								Field: heistv1alpha1.VaultBindingCertificateFieldTypeCertificate,
							},
						},
						"tls.key": {
							Certificate: &heistv1alpha1.VaultSyncCertificateSource{
								Name:  "template_0",
								Field: heistv1alpha1.VaultBindingCertificateFieldTypePrivateKey,
							},
						},
					},
				},
			}
			Test.K8sEnv.Create(sync)

			Test.K8sEnv.Object(sync).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been synced",
			))

			secret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      sync.Spec.Target.Name,
					Namespace: "default",
				},
			}
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)).To(Succeed())
		})

		AfterEach(func() {
			Test.K8sEnv.CleanupCreatedObject()
		})

		It("should have correctly synced the secret", func() {
			Expect(secret.Data).To(HaveLen(9))
			Expect(secret.Data["full_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["cert_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["cert"]).NotTo(BeEmpty())
			Expect(secret.Data["key"]).NotTo(BeEmpty())
			Expect(secret.Data["issued_full_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["issued_cert_chain"]).NotTo(BeEmpty())
			Expect(secret.Data["tls.crt"]).NotTo(BeEmpty())
			Expect(secret.Data["tls.key"]).NotTo(BeEmpty())
			Expect(secret.Data["some_key"]).To(Equal([]byte("ASDF ASDF")))

			publicInfo, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAInfoSecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(publicInfo).NotTo(BeNil())
			privateInfo, err := Test.RootAPI.ReadKvSecret(common.InternalKvEngine, core.SecretPath(common.GetCAPrivateKeySecretPath(intermediateCA)))
			Expect(err).NotTo(HaveOccurred())
			Expect(privateInfo).NotTo(BeNil())

			Expect(publicInfo.Fields["full_certificate_chain"]).To(Equal(string(secret.Data["full_chain"])))
			Expect(publicInfo.Fields["certificate_chain"]).To(Equal(string(secret.Data["cert_chain"])))
			Expect(publicInfo.Fields["certificate"]).To(Equal(string(secret.Data["cert"])))
			Expect(privateInfo.Fields["private_key"]).To(Equal(string(secret.Data["key"])))
		})

		It("should have correctly added label and annotation to target secret", func() {
			Expect(secret.Data).To(HaveLen(9))

			// More complex logic needed to correctly update syncsecret spec without
			// running into race condition with controller. The reason for this
			// is that the controller is still in the process of updating the
			// fields in the secret, while we are already fetching it and
			// setting the path.
			updateSyncSecret := func() (err error) {
				resourceWithLabelsAndAnnotations := sync.DeepCopy()

				err = Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(resourceWithLabelsAndAnnotations), resourceWithLabelsAndAnnotations)
				if err != nil {
					return fmt.Errorf("couldn't fetch current syncsecret from apiserver: %w", err)
				}

				resourceWithLabelsAndAnnotations.Spec.Target.AdditionalAnnotations = map[string]string{
					"youniqx.com/test-annotation": "true",
				}
				resourceWithLabelsAndAnnotations.Spec.Target.AdditionalLabels = map[string]string{
					"youniqx.com/test-label": "true",
				}

				err = Test.K8sClient.Update(context.TODO(), resourceWithLabelsAndAnnotations)
				if err != nil {
					return fmt.Errorf("couldn't update syncsecret in apiserver: %w", err)
				}

				err = Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(secret), secret)
				if err != nil {
					return fmt.Errorf("couldn't fetch current secret from apiserver: %w", err)
				}

				if secret.ObjectMeta.Annotations["youniqx.com/test-annotation"] != "true" {
					return fmt.Errorf("secret does not have annotation")
				}

				if secret.ObjectMeta.Labels["youniqx.com/test-label"] != "true" {
					return fmt.Errorf("secret does not have label")
				}

				return nil
			}

			Eventually(updateSyncSecret, 5*time.Second, 1*time.Second).Should(Succeed())

			Eventually(secret.ObjectMeta.Annotations["youniqx.com/test-annotation"]).Should(Equal("true"))
			Eventually(secret.ObjectMeta.Labels["youniqx.com/test-label"]).Should(Equal("true"))
		})

		It("should delete the secret if the sync object is deleted", func() {
			Expect(Test.K8sClient.Delete(context.TODO(), sync)).To(Succeed())
			Test.K8sEnv.Object(sync).Should(BeNil())
			Test.K8sEnv.Object(secret).Should(BeNil())
		})
	})
})
