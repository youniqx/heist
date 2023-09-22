package e2e_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
	. "github.com/youniqx/heist/pkg/vault/matchers"
)

var _ = Describe("Auth API", func() {
	When("Enabling new auth methods", func() {
		k8sAuthMethod := &auth.Method{
			Path: "managed/k8s",
			Type: auth.MethodKubernetes,
		}

		AfterEach(func() {
			Expect(vaultAPI.DeleteAuthMethod(k8sAuthMethod)).To(Succeed())
		})

		It("Should be able to enable a new kubernetes auth method", func() {
			Expect(vaultAPI.CreateAuthMethod(k8sAuthMethod)).To(Succeed())
			vaultEnv.AuthMethod(k8sAuthMethod).Should(HavePath("managed/k8s"))
			vaultEnv.AuthMethod(k8sAuthMethod).Should(HaveAuthType(auth.MethodKubernetes))
		})

		It("Should not list the auth methods if they have not been created yet", func() {
			methods, err := vaultAPI.ListAuthMethods()
			Expect(err).NotTo(HaveOccurred())
			Expect(methods).NotTo(ContainElements(
				k8sAuthMethod,
			))
		})
	})

	When("Managing existing auth methods", func() {
		authMethod := &auth.Method{
			Path: "managed/k8s",
			Type: auth.MethodKubernetes,
		}

		BeforeEach(func() {
			Expect(vaultAPI.CreateAuthMethod(authMethod)).To(Succeed())
		})

		AfterEach(func() {
			Expect(vaultAPI.DeleteAuthMethod(authMethod)).To(Succeed())
		})

		It("Should throw an error when trying to create an auth method at a path that is already in use", func() {
			Expect(vaultAPI.CreateAuthMethod(authMethod)).NotTo(Succeed())
		})

		It("Should be able to read an existing auth method", func() {
			method, err := vaultAPI.ReadAuthMethod(authMethod)
			Expect(err).NotTo(HaveOccurred())
			Expect(method).To(Equal(authMethod))
		})

		It("Should throw an error when trying to read a non-existing auth method", func() {
			method, err := vaultAPI.ReadAuthMethod(core.MountPath("does/not/exist"))
			Expect(err).To(MatchError(core.ErrDoesNotExist))
			Expect(method).To(BeNil())
		})

		It("Should include the existing auth methods when listing auth methods", func() {
			methods, err := vaultAPI.ListAuthMethods()
			Expect(err).NotTo(HaveOccurred())
			Expect(methods).To(ContainElement(authMethod))
		})

		It("Should return true when checking if an existing auth method exists", func() {
			exists, err := vaultAPI.HasAuthMethod(authMethod)
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("Should return false when checking if a non-existing auth method exists", func() {
			exists, err := vaultAPI.HasAuthMethod(core.MountPath("does/not/exist"))
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("Should be able to delete an existing auth method", func() {
			Expect(vaultAPI.DeleteAuthMethod(authMethod)).To(Succeed())
			vaultEnv.AuthMethod(authMethod).Should(BeNil())
		})

		It("Should not throw an error when trying to delete a non-existing auth method", func() {
			Expect(vaultAPI.DeleteAuthMethod(core.MountPath("does/not/exist"))).To(Succeed())
			vaultEnv.AuthMethod(core.MountPath("does/not/exist")).Should(BeNil())
		})
	})
})
