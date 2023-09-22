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

package agentserver_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/agent"
	"github.com/youniqx/heist/pkg/agentserver"
	"github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/vault/pki"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Heist Agent Server", func() {
	When("Starting the Agent Server", func() {
		var (
			engine           *v1alpha1.VaultKVSecretEngine
			secret           *v1alpha1.VaultKVSecret
			binding          *v1alpha1.VaultBinding
			serviceAccount   *v1.ServiceAccount
			instance         agentserver.Server
			secretOutputPath string
		)

		BeforeEach(func() {
			engine = &v1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-engine-21-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec:   v1alpha1.VaultKVSecretEngineSpec{},
				Status: v1alpha1.VaultKVSecretEngineStatus{},
			}

			secret = &v1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-secret-21-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*v1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: v1alpha1.EncryptedValue(DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: v1alpha1.VaultKVSecretStatus{},
			}

			serviceAccount = &v1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-service-account-21-%d", time.Now().Unix()),
					Namespace: "default",
				},
			}

			binding = &v1alpha1.VaultBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultBinding",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-binding-21-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultBindingSpec{
					Subject: v1alpha1.VaultBindingSubject{
						Name: serviceAccount.Name,
					},
					KVSecrets: []v1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					Agent: v1alpha1.VaultBindingAgentConfig{
						Templates: []v1alpha1.VaultBindingValueTemplate{
							{
								Path:     "some-field",
								Mode:     "0640",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
						},
					},
				},
			}

			var err error
			secretOutputPath, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(secretOutputPath).NotTo(BeEmpty())

			agentInstance, err := agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
				agent.WithBasePath(secretOutputPath),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(agentInstance).NotTo(BeNil())

			instance = agentserver.New(agentInstance)
			Expect(instance).NotTo(BeNil())
			go func() {
				defer GinkgoRecover()
				Expect(instance.ListenAndServer(":39704")).NotTo(Succeed())
			}()
			Eventually(instance.IsListening, 10*time.Second, 250*time.Millisecond).Should(BeTrue())
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()

			if instance != nil {
				instance.Stop()
			}
		})

		It("Should be in an unsynced state", func() {
			Expect(instance.IsSynced()).To(BeFalse())
		})

		It("Should eventually be able to sync all secrets to disk after the agent could sync the config", func() {
			K8sEnv.Create(engine, secret, binding, serviceAccount)

			Eventually(instance.IsSynced, 30*time.Second, 250*time.Millisecond).Should(BeTrue())

			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "some-field"))).To(Equal("ASDF ASDF"))
			Expect(ReadFile(filepath.Join(secretOutputPath, "config.json"))).NotTo(BeEmpty())
		})
	})

	When("Running a synced Agent Server", func() {
		var (
			instance         agentserver.Server
			secretOutputPath string
		)

		BeforeEach(func() {
			engine := &v1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-engine-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec:   v1alpha1.VaultKVSecretEngineSpec{},
				Status: v1alpha1.VaultKVSecretEngineStatus{},
			}
			K8sEnv.Create(engine)

			secret := &v1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("kv-secret-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*v1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: v1alpha1.EncryptedValue(DefaultCipherText),
						},
						"some-field-2": {
							CipherText: v1alpha1.EncryptedValue(DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: v1alpha1.VaultKVSecretStatus{},
			}
			K8sEnv.Create(secret)

			rootCA := &v1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("root-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultCertificateAuthoritySpec{
					Subject: v1alpha1.VaultCertificateAuthoritySubject{
						CommonName: "my-root",
					},
					Settings: v1alpha1.VaultCertificateAuthoritySettings{
						KeyType: pki.KeyTypeRSA,
						KeyBits: pki.KeyBitsRSA2048,
					},
				},
			}
			K8sEnv.Create(rootCA)

			intermediateCA := &v1alpha1.VaultCertificateAuthority{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("intermediate-ca-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultCertificateAuthoritySpec{
					Issuer: rootCA.Name,
					Subject: v1alpha1.VaultCertificateAuthoritySubject{
						CommonName: "my-intermediate",
					},
					Settings: v1alpha1.VaultCertificateAuthoritySettings{
						KeyType:  pki.KeyTypeRSA,
						KeyBits:  pki.KeyBitsRSA2048,
						Exported: true,
					},
				},
			}
			K8sEnv.Create(intermediateCA)

			certificate := &v1alpha1.VaultCertificateRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("cert-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultCertificateRoleSpec{
					Issuer: intermediateCA.Name,
					Settings: v1alpha1.VaultCertificateRoleSettings{
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

			binding := &v1alpha1.VaultBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("some-binding-12-%d", time.Now().Unix()),
					Namespace: "default",
				},
				Spec: v1alpha1.VaultBindingSpec{
					Subject: v1alpha1.VaultBindingSubject{
						Name: serviceAccount.Name,
					},
					KVSecrets: []v1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					CertificateAuthorities: []v1alpha1.VaultBindingCertificateAuthority{
						{
							Name: intermediateCA.Name,
							Capabilities: []v1alpha1.VaultBindingCertificateAuthorityCapability{
								v1alpha1.VaultBindingCertificateAuthorityCapabilityReadPrivate,
							},
						},
					},
					CertificateRoles: []v1alpha1.VaultBindingCertificate{
						{
							Name: certificate.Name,
							Capabilities: []v1alpha1.VaultBindingCertificateCapability{
								v1alpha1.VaultBindingCertificateCapabilityIssue,
							},
						},
					},
					Agent: v1alpha1.VaultBindingAgentConfig{
						Templates: []v1alpha1.VaultBindingValueTemplate{
							{
								Path:     "some-field",
								Mode:     "0755",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
							{
								Path:     "some-field-2",
								Mode:     "0644",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field-2\" }}", secret.Name),
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
								Template: "{{ certField \"cert_template_001\" \"private_key\" }}",
							},
							{
								Path:     "certificate",
								Template: "{{ certField \"cert_template_001\" \"certificate\" }}",
							},
							{
								Path:     "certificate-chain",
								Template: "{{ certField \"cert_template_001\" \"cert_chain\" }}",
							},
						},
						CertificateTemplates: []v1alpha1.VaultCertificateTemplate{
							{
								Alias:             "cert_template_001",
								CertificateRole:   certificate.Name,
								CommonName:        "some-common-name",
								ExcludeCNFromSans: true,
							},
						},
					},
				},
				Status: v1alpha1.VaultBindingStatus{},
			}
			K8sEnv.Create(binding)

			var err error
			secretOutputPath, err = os.MkdirTemp("", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(secretOutputPath).NotTo(BeEmpty())

			agentInstance, err := agent.New(
				agent.WithRestConfig(AgentConfig),
				agent.WithVaultToken(VaultEnv.GetRootToken()),
				agent.WithClientConfig("default", serviceAccount.Name),
				agent.WithBasePath(secretOutputPath),
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(agentInstance).NotTo(BeNil())

			instance = agentserver.New(agentInstance)
			go func() {
				defer GinkgoRecover()

				Expect(instance.ListenAndServer(":8082")).NotTo(Succeed())
			}()
			Expect(instance).NotTo(BeNil())

			Eventually(instance.IsSynced, 30*time.Second, 250*time.Millisecond).Should(BeTrue())
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()

			if instance != nil {
				instance.Stop()
			}
		})

		It("Should have written all secrets with the correct values to the disk", func() {
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "some-field"))).To(Equal("ASDF ASDF"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "some-field"))).To(Equal(os.FileMode(0o755)))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets"))).To(Equal(os.ModeDir | os.FileMode(0o755)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "some-field-2"))).To(Equal("ASDF ASDF"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "some-field-2"))).To(Equal(os.FileMode(0o644)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "intermediate-certificate"))).To(HavePrefix("-----"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "intermediate-certificate"))).To(Equal(os.FileMode(0o640)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "intermediate-private-key"))).To(HavePrefix("-----"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "intermediate-private-key"))).To(Equal(os.FileMode(0o640)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "certificate-private-key"))).To(HavePrefix("-----"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "certificate-private-key"))).To(Equal(os.FileMode(0o640)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "certificate"))).To(HavePrefix("-----"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "certificate"))).To(Equal(os.FileMode(0o640)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "secrets", "certificate-chain"))).To(HavePrefix("-----"))
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "secrets", "certificate-chain"))).To(Equal(os.FileMode(0o640)))
			Expect(ReadFile(filepath.Join(secretOutputPath, "config.json"))).NotTo(BeEmpty())
			Expect(ReadFilePerm(filepath.Join(secretOutputPath, "config.json"))).To(Equal(os.FileMode(0o640)))
		})
	})
})

func ReadFile(path string) string {
	data, err := ioutil.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	return string(data)
}

func ReadFilePerm(path string) os.FileMode {
	data, err := os.Stat(path)
	Expect(err).NotTo(HaveOccurred())
	return data.Mode()
}

type mockAgent struct {
	BasePath string
	Synced   bool
}

func (m *mockAgent) GetClientSecret() *agent.Secret {
	return &agent.Secret{
		Value:      "some config",
		Name:       "heist.json",
		OutputPath: filepath.Join(m.BasePath, "config.json"),
		Mode:       0o644,
	}
}

func (m *mockAgent) CreateUpdateChannel(bools chan bool) {
}

func (m *mockAgent) ListSecrets() (names []string, err error) {
	if !m.Synced {
		return nil, agent.ErrNotYetSynced
	}

	return []string{"some-secret"}, nil
}

func (m *mockAgent) FetchSecret(name string) (secret *agent.Secret, err error) {
	if !m.Synced {
		return nil, agent.ErrNotYetSynced
	}

	switch name {
	case "some-secret":
		return &agent.Secret{
			Value:      "ASDF ASDF",
			Name:       name,
			OutputPath: filepath.Join(m.BasePath, "secrets", name),
			Mode:       0o640,
		}, nil
	default:
		return nil, agent.ErrNotFound
	}
}

func (m *mockAgent) GetStatus() *agent.SyncStatus {
	if !m.Synced {
		return &agent.SyncStatus{
			Status: agent.StatusNotYetSynced,
			Reason: "not yet synced",
		}
	}

	return &agent.SyncStatus{
		Status: agent.StatusSynced,
		Reason: "All secrets have been synced",
	}
}

func (m *mockAgent) Stop() {
}

var _ = Describe("The Agent Server REST API", func() {
	var (
		recorder         *httptest.ResponseRecorder
		instance         agentserver.Server
		secretOutputPath string
	)

	BeforeEach(func() {
		var err error
		secretOutputPath, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())
		Expect(secretOutputPath).NotTo(BeEmpty())
		recorder = httptest.NewRecorder()
	})

	Context("If the Agent Server is not synced", func() {
		BeforeEach(func() {
			instance = agentserver.New(&mockAgent{
				BasePath: secretOutputPath,
				Synced:   false,
			})
			go func() {
				defer GinkgoRecover()
				Expect(instance.ListenAndServer(":8083")).NotTo(Succeed())
			}()
			Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeTrue())
			Eventually(instance.IsSynced, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
		})

		AfterEach(func() {
			instance.Stop()
		})

		When("interacting with the liveness probe at /live", func() {
			Context("If making a GET request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("GET", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("POST", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})
		})

		When("interacting with the readiness probe at /ready", func() {
			Context("If making a GET request", func() {
				It("returns a 500 status code", func() {
					request, _ := http.NewRequest("GET", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 500 status code", func() {
					request, _ := http.NewRequest("POST", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})
		})

		When("sending a shutdown request to /shutdown", func() {
			Context("If making a GET request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("GET", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("POST", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})
			})
		})
	})

	Context("If the Agent Server is synced", func() {
		BeforeEach(func() {
			instance = agentserver.New(&mockAgent{
				BasePath: secretOutputPath,
				Synced:   true,
			})
			go func() {
				defer GinkgoRecover()
				Expect(instance.ListenAndServer(":8085")).NotTo(Succeed())
			}()
			Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeTrue())
			Eventually(instance.IsSynced, 10*time.Second, 200*time.Millisecond).Should(BeTrue())
		})

		AfterEach(func() {
			instance.Stop()
		})

		When("interacting with the liveness probe at /live", func() {
			Context("If making a GET request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("GET", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("POST", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/live", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})
		})

		When("interacting with the readiness probe at /ready", func() {
			Context("If making a GET request", func() {
				It("returns a 500 status code", func() {
					request, _ := http.NewRequest("GET", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 500 status code", func() {
					request, _ := http.NewRequest("POST", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/ready", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
				})
			})
		})

		When("sending a shutdown request to /shutdown", func() {
			Context("If making a GET request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("GET", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("GET", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})
			})

			Context("If making a POST request", func() {
				It("returns a 200 status code", func() {
					request, _ := http.NewRequest("POST", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Code).To(Equal(http.StatusOK))
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})

				It("returns an empty body status code", func() {
					request, _ := http.NewRequest("POST", "/shutdown", nil)
					instance.ServeHTTP(recorder, request)
					Expect(recorder.Body.Bytes()).To(BeEmpty())
					Eventually(instance.IsListening, 10*time.Second, 200*time.Millisecond).Should(BeFalse())
				})
			})
		})
	})
})
