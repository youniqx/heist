package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/auth"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/kubernetesauth"
	. "github.com/youniqx/heist/pkg/vault/matchers"
)

var _ = Describe("KubernetesRolesAPI", func() {
	It("Should be able to manage kubernetes auth roles", func() {
		By("Creating a new kubernetes auth method")
		method := &kubernetesauth.Method{
			Path: "managed/test-auth",
			Config: &kubernetesauth.Config{
				KubernetesHost:       "https://kubernetes.default.svc.cluster.local",
				Issuer:               "https://kubernetes.default.svc.cluster.local",
				PemKeys:              nil,
				KubernetesCACert:     "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAoGdpfQA9y+2KP2rwyeZD\nwW9o8yWJQpSSnVr2thpdxYSO4U4Phv+DvgW8LIu8+nRUzX1GbyIaGIqSDEOdJsEE\ne8EywIxM4TWTXnu3mf9lxYx3+RLqirmGKecDZ17m5gRnnokjmoJi3GoJOYzlrfBC\n3srBVwew5d/yrFEBPdM/JHnfE+J6J5HASThrF/WxNjR/HrUREgUUGxxfj0OkCJqX\njs2Fm24jSZzummSCHlzxOh/jWcZgvWuOUi+LauKOXQcc7HQcMgakrnfGHyGqIxVY\nX/C7CPynMWzkmaf9SxsrdgCC6eJS8VqyCq2qk4T2Oyvcvfxg7JBmxyHzmmNwyoYK\njwIDAQAB\n-----END PUBLIC KEY-----\n",
				TokenReviewerJWT:     "",
				DisableISSValidation: false,
				DisableLocalCAJWT:    false,
			},
		}
		Expect(vaultAPI.UpdateKubernetesAuthMethod(method)).To(Succeed())
		vaultEnv.AuthMethod(method).Should(HavePath("managed/test-auth"))
		vaultEnv.AuthMethod(method).Should(HaveAuthType(auth.MethodKubernetes))

		By("Creating a role")
		role := &kubernetesauth.Role{
			Name:                 "some-name",
			BoundNamespaces:      []string{"some-namespace"},
			BoundServiceAccounts: []string{"some-service-account"},
			Policies: []core.PolicyName{
				"some-policy",
				"some-other-policy",
			},
		}
		Expect(vaultAPI.UpdateKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(HaveName("some-name"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToNamespaces("some-namespace"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToServiceAccounts("some-service-account"))
		vaultEnv.KubernetesAuthRole(method, role).Should(HavePolicies(
			core.PolicyName("some-policy"),
			core.PolicyName("some-other-policy"),
		))

		By("Not making an update if nothing has changed")
		Expect(vaultAPI.UpdateKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(HaveName("some-name"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToNamespaces("some-namespace"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToServiceAccounts("some-service-account"))
		vaultEnv.KubernetesAuthRole(method, role).Should(HavePolicies(
			core.PolicyName("some-policy"),
			core.PolicyName("some-other-policy"),
		))

		By("Updating it if policy changes")
		role.Policies[1] = "some-new-policy"
		Expect(vaultAPI.UpdateKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(HaveName("some-name"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToNamespaces("some-namespace"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToServiceAccounts("some-service-account"))
		vaultEnv.KubernetesAuthRole(method, role).Should(HavePolicies(
			core.PolicyName("some-policy"),
			core.PolicyName("some-new-policy"),
		))

		By("Updating it if serviceAccount changes")
		role.BoundServiceAccounts[0] = "some-new-service-account"
		Expect(vaultAPI.UpdateKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(HaveName("some-name"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToNamespaces("some-namespace"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToServiceAccounts("some-new-service-account"))
		vaultEnv.KubernetesAuthRole(method, role).Should(HavePolicies(
			core.PolicyName("some-policy"),
			core.PolicyName("some-new-policy"),
		))

		By("Updating it if serviceAccountNamespace changes")
		role.BoundNamespaces[0] = "some-new-namespace"
		Expect(vaultAPI.UpdateKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(HaveName("some-name"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToNamespaces("some-new-namespace"))
		vaultEnv.KubernetesAuthRole(method, role).Should(BeBoundToServiceAccounts("some-new-service-account"))
		vaultEnv.KubernetesAuthRole(method, role).Should(HavePolicies(
			core.PolicyName("some-policy"),
			core.PolicyName("some-new-policy"),
		))

		By("Deleting it")
		Expect(vaultAPI.DeleteKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(BeNil())

		By("Not throwing an error when trying to delete a non existent one")
		Expect(vaultAPI.DeleteKubernetesAuthRole(method, role)).To(Succeed())
		vaultEnv.KubernetesAuthRole(method, role).Should(BeNil())
	})
})
