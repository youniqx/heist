package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	"github.com/youniqx/heist/pkg/vault/kvsecret"
	. "github.com/youniqx/heist/pkg/vault/matchers"
)

var _ = Describe("KV Secret API", func() {
	When("Creating a new Secret", func() {
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

		BeforeEach(func() {
			Expect(vaultAPI.UpdateKvEngine(engine)).Should(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
		})

		It("Should not exist before creating it", func() {
			vaultEnv.KvSecret(engine, secret).Should(BeNil())
		})

		It("Should be able to create a secret in an existing engine", func() {
			Expect(vaultAPI.UpdateKvSecret(engine, secret)).To(Succeed())
			vaultEnv.KvSecret(engine, secret).Should(HavePath("some-secret"))
			vaultEnv.KvSecret(engine, secret).Should(HaveKvSecretFields(map[string]string{
				"some-field": "some-value",
			}))
		})

		It("Should throw an error if the engine does not exist", func() {
			Expect(vaultAPI.UpdateKvSecret(core.MountPath("does/not/exist"), secret)).NotTo(Succeed())
			vaultEnv.KvSecret(engine, secret).Should(BeNil())
		})
	})

	When("Managing existing secrets", func() {
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

		BeforeEach(func() {
			Expect(vaultAPI.UpdateKvEngine(engine)).Should(Succeed())
			Expect(vaultAPI.UpdateKvSecret(engine, secret)).Should(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).Should(Succeed())
		})

		It("Should be able to read an existing secret", func() {
			info, err := vaultAPI.ReadKvSecret(engine, secret)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(secret))
		})

		It("Should throw an error when reading a non-existing secret", func() {
			info, err := vaultAPI.ReadKvSecret(engine, core.SecretPath("does/not/exist"))
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(info).To(BeNil())
		})

		It("Should throw an error when reading a secret in a non existing engine", func() {
			info, err := vaultAPI.ReadKvSecret(core.MountPath("does/not/exist"), secret)
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(info).To(BeNil())
		})

		It("Should be able to delete an existing secret", func() {
			Expect(vaultAPI.DeleteKvSecret(engine, secret)).To(Succeed())
			vaultEnv.KvSecret(engine, secret).Should(BeNil())
		})

		It("Should not throw an error when deleting a non-existing secret in an existing engine", func() {
			Expect(vaultAPI.DeleteKvSecret(engine, core.SecretPath("does/not/exist"))).To(Succeed())
			vaultEnv.KvSecret(engine, core.SecretPath("does/not/exist")).Should(BeNil())
		})

		It("Should not throw an error when deleting an existing secret in a non-existing engine", func() {
			Expect(vaultAPI.DeleteKvSecret(core.MountPath("does/not/exist"), secret)).To(Succeed())
			vaultEnv.KvSecret(core.MountPath("does/not/exist"), secret).Should(BeNil())
		})
	})
})
