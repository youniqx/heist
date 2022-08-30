package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/core"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	"github.com/youniqx/heist/pkg/vault/transit"
)

var _ = Describe("Transit API", func() {
	When("Creating a new transit engine", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should create a new transit engine with the desired config", func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
			vaultEnv.TransitEngine(engine).Should(HavePath("some/path"))
			vaultEnv.TransitEngine(engine).Should(HaveConfig(engine.Config))
		})
	})

	When("Managing an empty transit engine", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should have created the engine with the desired config", func() {
			vaultEnv.TransitEngine(engine).Should(HavePath("some/path"))
			vaultEnv.TransitEngine(engine).Should(HaveConfig(engine.Config))
		})

		It("Should be able to read from an existing engine", func() {
			info, err := vaultAPI.ReadTransitEngine(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(engine))
		})

		It("Should throw an error when trying to read from a non-existing engine", func() {
			info, err := vaultAPI.ReadTransitEngine(core.MountPath("does/not/exist"))
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(info).To(BeNil())
		})

		It("Should return an empty list when trying to list keys", func() {
			keys, err := vaultAPI.ListKeys(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(keys).To(BeEmpty())
		})
	})

	When("Creating an encryption key", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		key := &transit.Key{
			Name: "some-key",
			Type: transit.TypeAes256Gcm96,
			Config: &transit.KeyConfig{
				MinimumDecryptionVersion: 1,
				MinimumEncryptionVersion: 1,
				DeletionAllowed:          false,
				Exportable:               false,
				AllowPlaintextBackup:     false,
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should have created the engine with the desired config", func() {
			vaultEnv.TransitEngine(engine).Should(HavePath("some/path"))
			vaultEnv.TransitEngine(engine).Should(HaveConfig(engine.Config))
		})

		It("Should be able to create a new encryption key", func() {
			Expect(vaultAPI.UpdateTransitKey(engine, key)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(key.Config))
		})

		It("Should be able to create a new encryption key", func() {
			Expect(vaultAPI.UpdateTransitKey(engine, key)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(key.Config))
		})
	})

	When("Managing an existing encryption key", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		key := &transit.Key{
			Name: "some-key",
			Type: transit.TypeAes256Gcm96,
			Config: &transit.KeyConfig{
				MinimumDecryptionVersion: 1,
				MinimumEncryptionVersion: 1,
				DeletionAllowed:          false,
				Exportable:               false,
				AllowPlaintextBackup:     false,
			},
		}

		BeforeEach(func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
			Expect(vaultAPI.UpdateTransitKey(engine, key)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should have created the engine and key with the desired config", func() {
			vaultEnv.TransitEngine(engine).Should(HavePath(engine.Path))
			vaultEnv.TransitEngine(engine).Should(HaveConfig(engine.Config))
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(key.Config))
		})

		It("Should be able to list the encryption key", func() {
			keys, err := vaultAPI.ListKeys(engine)
			Expect(err).NotTo(HaveOccurred())
			Expect(keys).To(Equal([]transit.KeyName{
				transit.KeyName(key.Name),
			}))
		})

		It("Should be able to read the existing key", func() {
			info, err := vaultAPI.ReadTransitKey(engine, key)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(Equal(key))
		})

		It("Should throw an error when trying to read a non-existing key", func() {
			info, err := vaultAPI.ReadTransitKey(engine, transit.KeyName("does-not-exist"))
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(info).To(BeNil())
		})

		It("Should be able to rotate the encryption key", func() {
			Expect(vaultAPI.RotateTransitKey(engine, key)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(key.Config))
		})

		It("Should be able to change the key configuration", func() {
			keyWithDifferentConfig := &transit.Key{
				Name: key.Name,
				Type: key.Type,
				Config: &transit.KeyConfig{
					MinimumDecryptionVersion: 1,
					MinimumEncryptionVersion: 1,
					DeletionAllowed:          true,
					Exportable:               false,
					AllowPlaintextBackup:     false,
				},
			}
			Expect(vaultAPI.UpdateTransitKey(engine, keyWithDifferentConfig)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(keyWithDifferentConfig.Config))
		})

		It("Should throw an error when trying to delete a key which has DeletionAllowed set to false", func() {
			Expect(vaultAPI.DeleteTransitKey(engine, key)).NotTo(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(key.Config))
		})

		It("Should be able to delete a key after setting the DeletionAllowed flag", func() {
			keyWithDeletionAllowed := &transit.Key{
				Name: key.Name,
				Type: key.Type,
				Config: &transit.KeyConfig{
					MinimumDecryptionVersion: 1,
					MinimumEncryptionVersion: 1,
					DeletionAllowed:          true,
					Exportable:               false,
					AllowPlaintextBackup:     false,
				},
			}
			Expect(vaultAPI.UpdateTransitKey(engine, keyWithDeletionAllowed)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(HaveName(key.Name))
			vaultEnv.TransitKey(engine, key).Should(HaveKeyType(key.Type))
			vaultEnv.TransitKey(engine, key).Should(HaveConfig(keyWithDeletionAllowed.Config))
			Expect(vaultAPI.DeleteTransitKey(engine, key)).To(Succeed())
			vaultEnv.TransitKey(engine, key).Should(BeNil())
		})
	})

	When("Using a symmetric encryption key", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		key := &transit.Key{
			Name: "some-key",
			Type: transit.TypeAes256Gcm96,
			Config: &transit.KeyConfig{
				MinimumDecryptionVersion: 1,
				MinimumEncryptionVersion: 1,
				DeletionAllowed:          false,
				Exportable:               false,
				AllowPlaintextBackup:     false,
			},
		}

		inputPlainText := []byte("ASDF ASDF")

		BeforeEach(func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
			Expect(vaultAPI.UpdateTransitKey(engine, key)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should be able to encrypt and decrypt a value using the encryption key", func() {
			cipherText, err := vaultAPI.TransitEncrypt(engine, key, inputPlainText)
			Expect(err).NotTo(HaveOccurred())
			Expect(cipherText).NotTo(BeEmpty())
			plainText, err := vaultAPI.TransitDecrypt(engine, key, cipherText)
			Expect(err).NotTo(HaveOccurred())
			Expect(plainText).To(Equal(inputPlainText))
		})

		It("Should throw an error when trying to sign any input", func() {
			signature, err := vaultAPI.TransitSign(engine, key, inputPlainText)
			Expect(err).To(HaveOccurred())
			Expect(signature).To(BeEmpty())
		})

		It("Should throw an error when trying to validate any signature", func() {
			valid, err := vaultAPI.TransitVerify(engine, key, inputPlainText, "vault:v1:aaaaaaaaa")
			Expect(err).To(HaveOccurred())
			Expect(valid).To(BeFalse())
		})
	})

	When("Using an asymmetric encryption key", func() {
		engine := &transit.Engine{
			Path: "some/path",
			Config: &transit.EngineConfig{
				Cache: transit.EngineCacheConfig{
					Size: 1024,
				},
			},
		}

		key := &transit.Key{
			Name: "some-key",
			Type: transit.TypeRSA2048,
			Config: &transit.KeyConfig{
				MinimumDecryptionVersion: 1,
				MinimumEncryptionVersion: 1,
				DeletionAllowed:          false,
				Exportable:               false,
				AllowPlaintextBackup:     false,
			},
		}

		inputPlainText := []byte("ASDF ASDF")

		BeforeEach(func() {
			Expect(vaultAPI.UpdateTransitEngine(engine)).To(Succeed())
			Expect(vaultAPI.UpdateTransitKey(engine, key)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(engine)).To(Succeed())
		})

		It("Should be able to encrypt and decrypt a value using the encryption key", func() {
			cipherText, err := vaultAPI.TransitEncrypt(engine, key, inputPlainText)
			Expect(err).NotTo(HaveOccurred())
			Expect(cipherText).NotTo(BeEmpty())
			plainText, err := vaultAPI.TransitDecrypt(engine, key, cipherText)
			Expect(err).NotTo(HaveOccurred())
			Expect(plainText).To(Equal(inputPlainText))
		})

		It("Should be able to sign any input and verify those signatures", func() {
			signature, err := vaultAPI.TransitSign(engine, key, inputPlainText)
			Expect(err).NotTo(HaveOccurred())
			Expect(signature).NotTo(BeEmpty())
			valid, err := vaultAPI.TransitVerify(engine, key, inputPlainText, signature)
			Expect(err).NotTo(HaveOccurred())
			Expect(valid).To(BeTrue())
		})
	})
})
