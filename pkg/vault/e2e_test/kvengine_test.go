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

package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	. "github.com/youniqx/heist/pkg/vault/matchers"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("KV Engine API", func() {
	When("Creating a KV Secret Engine", func() {
		engine := &kvengine.KvEngine{
			Path: "managed/kv/some-engine",
			Config: &kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			},
		}
		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
		})

		It("Should not exist initially", func() {
			vaultEnv.KvEngine(engine).Should(BeNil())
		})

		It("Should be created with the intended config", func() {
			Expect(vaultAPI.UpdateKvEngine(engine)).To(Succeed())
			vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			}))
		})

		It("Should throw an error when trying to read the kv engine before creating it is completed", func() {
			info, err := vaultAPI.ReadKvEngine(engine)
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(info).To(BeNil())
		})
	})

	When("Managing an empty KV Secret Engine", func() {
		engine := &kvengine.KvEngine{
			Path: "managed/kv/some-engine",
			Config: &kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			},
		}
		BeforeEach(func() {
			Expect(vaultAPI.UpdateKvEngine(engine)).To(Succeed())
			vaultEnv.KvEngine(engine).ShouldNot(BeNil())
			vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			}))
		})
		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should have the intended config", func() {
			vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			}))
		})

		It("Should be possible to read the kv engine", func() {
			info, err := vaultAPI.ReadKvEngine(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(&kvengine.KvEngine{
				Path:   engine.Path,
				Config: engine.Config,
			}))
		})

		It("Should be able to change the config", func() {
			engineWithDifferentConfig := &kvengine.KvEngine{
				Path: engine.Path,
				Config: &kvengine.Config{
					MaxVersions:        20,
					CasRequired:        false,
					DeleteVersionAfter: "0s",
				},
			}
			Expect(vaultAPI.UpdateKvEngine(engineWithDifferentConfig)).To(Succeed())
			vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
				MaxVersions:        20,
				CasRequired:        false,
				DeleteVersionAfter: "0s",
			}))
		})

		It("Should list zero secrets", func() {
			secrets, err := vaultAPI.ListKvSecrets(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(secrets).To(HaveLen(0))
		})

		It("Should be able to be deleted", func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
			vaultEnv.KvEngine(engine).Should(BeNil())
		})

		It("Should not throw an error when deleting an already deleted engine", func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
			vaultEnv.KvEngine(engine).Should(BeNil())
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
		})
	})

	When("Managing a KV Secret Engine containing some secrets", func() {
		engine := &kvengine.KvEngine{
			Path: "managed/kv/some-engine",
			Config: &kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			},
		}
		secret := &kvsecret.KvSecret{
			Path: "some-secret",
			Fields: map[string]string{
				"some-field": "some-value",
			},
		}
		nestedSecret := &kvsecret.KvSecret{
			Path: "some/path/another-secret",
			Fields: map[string]string{
				"another-field": "another-value",
			},
		}
		BeforeEach(func() {
			Expect(vaultAPI.UpdateKvEngine(engine)).To(Succeed())
			Expect(vaultAPI.UpdateKvSecret(engine, secret)).To(Succeed())
			Expect(vaultAPI.UpdateKvSecret(engine, nestedSecret)).To(Succeed())
		})
		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should have the intended config", func() {
			vaultEnv.KvEngine(engine).Should(HavePath("managed/kv/some-engine"))
			vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			}))
		})

		It("Should contain the secrets with the correct fields", func() {
			vaultEnv.KvSecret(engine, secret).Should(HavePath("some-secret"))
			vaultEnv.KvSecret(engine, secret).Should(HaveKvSecretFields(map[string]string{
				"some-field": "some-value",
			}))
			vaultEnv.KvSecret(engine, nestedSecret).Should(HavePath("some/path/another-secret"))
			vaultEnv.KvSecret(engine, nestedSecret).Should(HaveKvSecretFields(map[string]string{
				"another-field": "another-value",
			}))
		})

		It("Should list the secrets it contains", func() {
			secrets, err := vaultAPI.ListKvSecrets(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(secrets).To(HaveLen(2))
			Expect(secrets).To(Equal([]core.SecretPath{
				"some-secret",
				"some/path/another-secret",
			}))
		})

		It("Should be able to be deleted", func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
			vaultEnv.KvEngine(engine).Should(BeNil())
		})

		It("Should delete the secret along with the engine", func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
			vaultEnv.KvEngine(engine).Should(BeNil())
			vaultEnv.KvSecret(engine, secret).Should(BeNil())
			vaultEnv.KvSecret(engine, nestedSecret).Should(BeNil())
		})
	})
})
