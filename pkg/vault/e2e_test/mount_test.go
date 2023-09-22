package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/core"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	"github.com/youniqx/heist/pkg/vault/mount"
)

var _ = Describe("Mount API", func() {
	When("Mounting new engines", func() {
		kvv1Engine := &mount.Mount{
			Path:   "some/path",
			Type:   mount.TypeKVV1,
			Config: nil,
		}

		kvv2Engine := &mount.Mount{
			Path: "another/path",
			Type: mount.TypeKVV2,
			Options: map[string]string{
				"version": "2",
			},
		}

		transitEngine := &mount.Mount{
			Path:   "yet-another/path",
			Type:   mount.TypeTransit,
			Config: nil,
		}

		pkiEngine := &mount.Mount{
			Path:   "yet-another-new/path",
			Type:   mount.TypePKI,
			Config: nil,
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(kvv1Engine)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(kvv2Engine)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(transitEngine)).Should(Succeed())
		})

		It("Should be able to mount a new kv-v2 engine when setting version 2 explicitly", func() {
			Expect(vaultAPI.MountEngine(kvv2Engine)).Should(Succeed())
			vaultEnv.Mount(kvv2Engine).Should(HavePath("another/path"))
			vaultEnv.Mount(kvv2Engine).Should(HaveMountType(mount.TypeKVV2))
		})

		It("Should be able to mount a new kv-v2 engine when omitting the version 2 config", func() {
			engineWithoutConfig := &mount.Mount{
				Path:   kvv2Engine.Path,
				Type:   kvv2Engine.Type,
				Config: nil,
			}
			Expect(vaultAPI.MountEngine(engineWithoutConfig)).Should(Succeed())
			vaultEnv.Mount(engineWithoutConfig).Should(HavePath("another/path"))
			vaultEnv.Mount(engineWithoutConfig).Should(HaveMountType(mount.TypeKVV2))
		})

		It("Should be able to mount a new kv-v1 engine", func() {
			Expect(vaultAPI.MountEngine(kvv1Engine)).Should(Succeed())
			vaultEnv.Mount(kvv1Engine).Should(HavePath("some/path"))
			vaultEnv.Mount(kvv1Engine).Should(HaveMountType(mount.TypeKVV1))
		})

		It("Should be able to mount a new transit engine", func() {
			Expect(vaultAPI.MountEngine(transitEngine)).Should(Succeed())
			vaultEnv.Mount(transitEngine).Should(HavePath("yet-another/path"))
			vaultEnv.Mount(transitEngine).Should(HaveMountType(mount.TypeTransit))
		})

		It("Should be able to mount a new pki engine", func() {
			Expect(vaultAPI.MountEngine(pkiEngine)).Should(Succeed())
			vaultEnv.Mount(pkiEngine).Should(HavePath("yet-another-new/path"))
			vaultEnv.Mount(pkiEngine).Should(HaveMountType(mount.TypePKI))
		})

		It("Should be able to list preexisting mounts", func() {
			mounts, err := vaultAPI.ListMounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(mounts).NotTo(ContainElements(
				kvv1Engine,
				kvv2Engine,
				transitEngine,
			))
		})
	})

	When("Managing existing mounts", func() {
		kvEngine := &mount.Mount{
			Path: "some/path",
			Type: mount.TypeKVV2,
			Options: map[string]string{
				"version": "2",
			},
			Config: &mount.TuneConfig{
				MaxLeaseTTL:     core.NewTTL(0),
				DefaultLeaseTTL: core.NewTTL(0),
			},
		}

		transitEngine := &mount.Mount{
			Path: "another/path",
			Type: mount.TypeTransit,
			Config: &mount.TuneConfig{
				MaxLeaseTTL:     core.NewTTL(0),
				DefaultLeaseTTL: core.NewTTL(0),
			},
		}

		tuneConfig := &mount.TuneConfig{
			DefaultLeaseTTL: core.NewTTL(3 * core.Day),
			MaxLeaseTTL:     core.NewTTL(4 * core.Week),
		}

		BeforeEach(func() {
			Expect(vaultAPI.MountEngine(kvEngine)).Should(Succeed())
			Expect(vaultAPI.MountEngine(transitEngine)).Should(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteEngine(kvEngine)).Should(Succeed())
			Expect(vaultAPI.DeleteEngine(transitEngine)).Should(Succeed())
		})

		It("Should include the mounted engine in the returned mount list", func() {
			mounts, err := vaultAPI.ListMounts()
			Expect(err).NotTo(HaveOccurred())
			Expect(mounts).To(ContainElements(
				kvEngine,
				transitEngine,
			))
		})

		It("Should throw an error when trying to mount something at a path that is already in use", func() {
			Expect(vaultAPI.MountEngine(kvEngine)).NotTo(Succeed())
		})

		It("Should be able to tune the engine", func() {
			Expect(vaultAPI.TuneEngine(kvEngine, tuneConfig)).To(Succeed())
			vaultEnv.TuneConfig(kvEngine).Should(Equal(tuneConfig))
		})

		It("Should be able read the engine tune config", func() {
			tuneConfig, err := vaultAPI.ReadTuneConfig(kvEngine)
			Expect(tuneConfig).To(Equal(&mount.TuneConfig{
				DefaultLeaseTTL: core.NewTTL(32 * core.Day),
				MaxLeaseTTL:     core.NewTTL(32 * core.Day),
			}))
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return true when checking if an existing mount exists", func() {
			exists, err := vaultAPI.HasEngine(kvEngine)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("Should be able to read an existing mount", func() {
			engine, err := vaultAPI.ReadMount(kvEngine)
			Expect(err).NotTo(HaveOccurred())
			Expect(engine).To(Equal(kvEngine))
		})

		It("Should be able to delete an existing mount", func() {
			Expect(vaultAPI.DeleteEngine(kvEngine)).To(Succeed())
			vaultEnv.Mount(kvEngine).Should(BeNil())
		})

		It("Should return false when checking if a non-existing mount exists", func() {
			exists, err := vaultAPI.HasEngine(core.MountPath("some/non-existing/engine"))
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("Should throw an error when trying to read a non-existing mount", func() {
			engine, err := vaultAPI.ReadMount(core.MountPath("some/non-existing/engine"))
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(engine).To(BeNil())
		})

		It("Should not throw an error when trying to delete a non-existing mount", func() {
			Expect(vaultAPI.DeleteEngine(core.MountPath("some/non-existing/engine"))).To(Succeed())
		})

		It("Should be able to reload the transit plugin backend", func() {
			Expect(vaultAPI.ReloadPluginBackends(mount.PluginTransit)).To(Succeed())
		})
	})
})
