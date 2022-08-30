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

package agent_test

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/agent"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	. "github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault/pki"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Heist Agent", func() {
	When("Starting the Agent", func() {
		var (
			engine         *heistv1alpha1.VaultKVSecretEngine
			secret         *heistv1alpha1.VaultKVSecret
			binding        *heistv1alpha1.VaultBinding
			serviceAccount *v1.ServiceAccount
			counter        int
			instance       agent.Agent
		)

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-engine-11-%d", counter),
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultKVSecretEngineSpec{},
				Status: heistv1alpha1.VaultKVSecretEngineStatus{},
			}

			secret = &heistv1alpha1.VaultKVSecret{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecret",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-secret-11-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: heistv1alpha1.EncryptedValue(DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: heistv1alpha1.VaultKVSecretStatus{},
			}

			serviceAccount = &v1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-service-account-11-%d", counter),
					Namespace: "default",
				},
			}

			binding = &heistv1alpha1.VaultBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultBinding",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-binding-11-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultBindingSpec{
					Subject: heistv1alpha1.VaultBindingSubject{
						Name: serviceAccount.Name,
					},
					KVSecrets: []heistv1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					Agent: heistv1alpha1.VaultBindingAgentConfig{
						Templates: []heistv1alpha1.VaultBindingValueTemplate{
							{
								Path:     "some-field",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
						},
					},
				},
			}
		})

		AfterEach(func() {
			if instance != nil {
				instance.Stop()
			}

			counter++
		})

		It("Should not start without a rest config", func() {
			instance, err := agent.New(
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
			)
			Expect(err).To(MatchError(agent.ErrInitAgentFailed))
			Expect(instance).To(BeNil())
		})

		It("Should not start without client config namespace and name being set", func() {
			instance, err := agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
			)
			Expect(err).To(MatchError(agent.ErrInitAgentFailed))
			Expect(instance).To(BeNil())
		})

		It("Should start properly with valid configuration", func() {
			var err error
			instance, err = agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())
			Expect(instance.GetStatus()).NotTo(BeNil())
			Expect(instance.GetStatus().Status).To(Equal(agent.StatusNotYetSynced))
		})

		It("Should sync after a while when the necessary resources were created", func() {
			var err error
			instance, err = agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())

			Expect(K8sClient.Create(context.TODO(), engine)).To(Succeed())
			Expect(K8sClient.Create(context.TODO(), secret)).To(Succeed())
			Expect(K8sClient.Create(context.TODO(), binding)).To(Succeed())
			Expect(K8sClient.Create(context.TODO(), serviceAccount)).To(Succeed())

			Eventually(func() agent.StatusType {
				return instance.GetStatus().Status
			}, 30*time.Second, 250*time.Millisecond).WithTimeout(3 * time.Minute).Should(Equal(agent.StatusSynced))

			Expect(K8sClient.Delete(context.TODO(), engine)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), secret)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), binding)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), serviceAccount)).To(Succeed())
		})
	})

	When("Running the Agent after it has synced", func() {
		var instance agent.Agent

		BeforeEach(func() {
			engine := &heistv1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-engine-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultKVSecretEngineSpec{},
				Status: heistv1alpha1.VaultKVSecretEngineStatus{},
			}
			K8sEnv.Create(engine)

			secret := &heistv1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-secret-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: heistv1alpha1.EncryptedValue(DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: heistv1alpha1.VaultKVSecretStatus{},
			}
			K8sEnv.Create(secret)

			rootCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						CommonName: "my-root",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						KeyType: pki.KeyTypeRSA,
						KeyBits: pki.KeyBitsRSA2048,
					},
				},
			}
			K8sEnv.Create(rootCA)

			intermediateCA := &heistv1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: heistv1alpha1.VaultCertificateAuthoritySubject{
						CommonName: "my-intermediate",
					},
					Settings: heistv1alpha1.VaultCertificateAuthoritySettings{
						KeyType:  pki.KeyTypeRSA,
						KeyBits:  pki.KeyBitsRSA2048,
						Exported: true,
					},
				},
			}
			K8sEnv.Create(intermediateCA)

			certificate := &heistv1alpha1.VaultCertificateRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("cert-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultCertificateRoleSpec{
					Issuer: intermediateCA.Name,
					Settings: heistv1alpha1.VaultCertificateRoleSettings{
						TTL:          metav1.Duration{Duration: time.Hour * 24},
						KeyType:      pki.KeyTypeRSA,
						KeyBits:      pki.KeyBitsRSA2048,
						AllowAnyName: true,
					},
				},
			}
			K8sEnv.Create(certificate)

			serviceAccount := &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-service-account-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
			}
			K8sEnv.Create(serviceAccount)

			binding := &heistv1alpha1.VaultBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-binding-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultBindingSpec{
					Subject: heistv1alpha1.VaultBindingSubject{
						Name: serviceAccount.Name,
					},
					Capabilities: nil,
					KVSecrets: []heistv1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					CertificateAuthorities: []heistv1alpha1.VaultBindingCertificateAuthority{
						{
							Name: intermediateCA.Name,
							Capabilities: []heistv1alpha1.VaultBindingCertificateAuthorityCapability{
								heistv1alpha1.VaultBindingCertificateAuthorityCapabilityReadPrivate,
							},
						},
					},
					CertificateRoles: []heistv1alpha1.VaultBindingCertificate{
						{
							Name: certificate.Name,
							Capabilities: []heistv1alpha1.VaultBindingCertificateCapability{
								heistv1alpha1.VaultBindingCertificateCapabilityIssue,
							},
						},
					},
					TransitKeys: nil,
					Agent: heistv1alpha1.VaultBindingAgentConfig{
						Templates: []heistv1alpha1.VaultBindingValueTemplate{
							{
								Path:     "some-field",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
							{
								Path:     "/vault/secrets/some-field",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
							{
								Path:     "intermediate-private-key",
								Template: fmt.Sprintf("{{ caField \"%s\" \"private_key\" }}", intermediateCA.Name),
							},
							{
								Path:     "intermediate-certificate",
								Template: fmt.Sprintf("{{ caField \"%s\" \"certificate\" }}", intermediateCA.Name),
							},
							{
								Path:     "certificate-private-key",
								Template: fmt.Sprintf("{{ certField \"%s\" \"private_key\" }}", certificate.Name),
							},
							{
								Path:     "certificate",
								Template: fmt.Sprintf("{{ certField \"%s\" \"certificate\" }}", certificate.Name),
							},
							{
								Path:     "certificate-chain",
								Template: fmt.Sprintf("{{ certField \"%s\" \"cert_chain\" }}", certificate.Name),
							},
						},
						CertificateTemplates: []heistv1alpha1.VaultCertificateTemplate{
							{
								Alias:           certificate.Name,
								CertificateRole: certificate.Name,
							},
						},
					},
				},
				Status: heistv1alpha1.VaultBindingStatus{},
			}
			K8sEnv.Create(binding)

			var err error
			instance, err = agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())
			Expect(instance.GetStatus()).NotTo(BeNil())
			Expect(instance.GetStatus().Status).To(Equal(agent.StatusNotYetSynced))

			K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))
			K8sEnv.Object(secret).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been provisioned",
			))
			K8sEnv.Object(rootCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			K8sEnv.Object(intermediateCA).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateAuthority has been provisioned",
			))
			K8sEnv.Object(certificate).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"CertificateRole has been provisioned",
			))
			K8sEnv.Object(binding).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Binding has been provisioned",
			))

			Eventually(func() agent.StatusType {
				return instance.GetStatus().Status
			}, 30*time.Second, 250*time.Millisecond).Should(Equal(agent.StatusSynced))
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()

			if instance != nil {
				instance.Stop()
			}
		})

		It("Should be able to list secrets", func() {
			secrets, err := instance.ListSecrets()
			Expect(err).NotTo(HaveOccurred())
			Expect(secrets).To(HaveLen(7))
			Expect(secrets).To(ContainElement("some-field"))
		})

		It("Should be able to return the client config secret", func() {
			config := instance.GetClientSecret()
			Expect(config).NotTo(BeNil())
			Expect(config.Name).To(Equal("heist.json"))
			Expect(config.Mode).To(Equal(os.FileMode(0o640)))
			Expect(config.OutputPath).To(Equal("/heist/config.json"))
			Expect(config.Value).NotTo(BeEmpty())
		})

		It("Should be able to fetch kv secrets", func() {
			secret, err := instance.FetchSecret("some-field")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(Equal("ASDF ASDF"))
			Expect(secret.Name).To(Equal("some-field"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/some-field"))
		})

		It("Should be able to fetch kv secrets with absolute paths", func() {
			secret, err := instance.FetchSecret("/vault/secrets/some-field")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(Equal("ASDF ASDF"))
			Expect(secret.Name).To(Equal("/vault/secrets/some-field"))
			Expect(secret.OutputPath).To(Equal("/vault/secrets/some-field"))
		})

		It("Should be able to fetch ca certificate", func() {
			secret, err := instance.FetchSecret("intermediate-certificate")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(HavePrefix("-----"))
			Expect(secret.Name).To(Equal("intermediate-certificate"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/intermediate-certificate"))
		})

		It("Should be able to fetch ca private key", func() {
			secret, err := instance.FetchSecret("intermediate-private-key")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(HavePrefix("-----BEGIN RSA PRIVATE KEY-----"))
			Expect(secret.Name).To(Equal("intermediate-private-key"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/intermediate-private-key"))
		})

		It("Should be able to fetch certificate private key", func() {
			secret, err := instance.FetchSecret("certificate-private-key")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(HavePrefix("-----"))
			Expect(secret.Name).To(Equal("certificate-private-key"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/certificate-private-key"))
		})

		It("Should be able to fetch certificate public key", func() {
			secret, err := instance.FetchSecret("certificate")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(HavePrefix("-----"))
			Expect(secret.Name).To(Equal("certificate"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/certificate"))
		})

		It("Should be able to fetch certificate chain", func() {
			secret, err := instance.FetchSecret("certificate-chain")
			Expect(err).NotTo(HaveOccurred())
			Expect(secret.Value).To(HavePrefix("-----"))
			Expect(secret.Name).To(Equal("certificate-chain"))
			Expect(secret.OutputPath).To(Equal("/heist/secrets/certificate-chain"))
		})
	})

	When("Running the Agent without the expected config existing", func() {
		var (
			engine         *heistv1alpha1.VaultKVSecretEngine
			secret         *heistv1alpha1.VaultKVSecret
			binding        *heistv1alpha1.VaultBinding
			serviceAccount *v1.ServiceAccount
			counter        int
			instance       agent.Agent
		)

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-engine-13-%d", counter),
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultKVSecretEngineSpec{},
				Status: heistv1alpha1.VaultKVSecretEngineStatus{},
			}

			secret = &heistv1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-secret-13-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: heistv1alpha1.EncryptedValue(DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: heistv1alpha1.VaultKVSecretStatus{},
			}

			serviceAccount = &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-service-account-13-%d", counter),
					Namespace: "default",
				},
			}

			binding = &heistv1alpha1.VaultBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultBinding",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-binding-13-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultBindingSpec{
					Subject: heistv1alpha1.VaultBindingSubject{
						Name: serviceAccount.Name,
					},
					Capabilities: nil,
					KVSecrets: []heistv1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					Agent: heistv1alpha1.VaultBindingAgentConfig{
						Templates: []heistv1alpha1.VaultBindingValueTemplate{
							{
								Path:     "some-field",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
						},
					},
				},
				Status: heistv1alpha1.VaultBindingStatus{},
			}

			K8sEnv.Create(engine)
			K8sEnv.Create(secret)
			K8sEnv.Create(binding)
			K8sEnv.Create(serviceAccount)

			K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))
			K8sEnv.Object(secret).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Secret has been provisioned",
			))
			K8sEnv.Object(binding).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Binding has been provisioned",
			))

			var err error
			instance, err = agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", "some-sa"),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(instance).NotTo(BeNil())
		})

		AfterEach(func() {
			Expect(K8sClient.Delete(context.TODO(), engine)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), secret)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), binding)).To(Succeed())
			Expect(K8sClient.Delete(context.TODO(), serviceAccount)).To(Succeed())

			if instance != nil {
				instance.Stop()
			}

			counter++
		})

		It("Should not be able to list secrets", func() {
			secrets, err := instance.ListSecrets()
			Expect(err).To(MatchError(agent.ErrNotYetSynced))
			Expect(secrets).To(BeNil())
		})

		It("Should not be able to fetch secrets", func() {
			value, err := instance.FetchSecret("some-field")
			Expect(err).To(MatchError(agent.ErrNotYetSynced))
			Expect(value).To(BeNil())
		})

		It("Should be in the not yet synced status", func() {
			Expect(instance.GetStatus().Status).To(Equal(agent.StatusNotYetSynced))
		})
	})
})
