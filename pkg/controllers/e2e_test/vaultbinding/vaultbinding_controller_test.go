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

package vaultbinding

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault/core"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VaultBinding Controller", func() {
	When("Creating a new VaultBinding", func() {
		var (
			engine  *heistv1alpha1.VaultKVSecretEngine
			secret  *heistv1alpha1.VaultKVSecret
			binding *heistv1alpha1.VaultBinding
			counter int
		)

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultKVSecretEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("engine-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretEngineSpec{},
			}
			Expect(Test.K8sClient.Create(context.TODO(), engine)).To(Succeed())

			secret = &heistv1alpha1.VaultKVSecret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("secret-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultKVSecretSpec{
					Engine: engine.Name,
					Fields: map[string]*heistv1alpha1.VaultKVSecretField{
						"some-field": {
							CipherText: heistv1alpha1.EncryptedValue(Test.DefaultCipherText),
						},
					},
					DeleteProtection: false,
				},
				Status: heistv1alpha1.VaultKVSecretStatus{},
			}
			Expect(Test.K8sClient.Create(context.TODO(), secret)).To(Succeed())
		})

		AfterEach(func() {
			Test.K8sEnv.DeleteIfPresent(engine, secret, binding)
			Test.K8sEnv.Object(engine).Should(BeNil())
			Test.K8sEnv.Object(secret).Should(BeNil())
			Test.K8sEnv.Object(binding).Should(BeNil())
			counter++
		})

		It("Should correctly create Role, RoleBinding and VaultClientConfig Objects", func() {
			saName := fmt.Sprintf("some-sa-%d", counter)
			binding = &heistv1alpha1.VaultBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("binding-%d", counter),
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultBindingSpec{
					Subject: heistv1alpha1.VaultBindingSubject{
						Name: saName,
					},
					KVSecrets: []heistv1alpha1.VaultBindingKV{
						{
							Name: secret.Name,
						},
					},
					Agent: heistv1alpha1.VaultBindingAgentConfig{
						Templates: []heistv1alpha1.VaultBindingValueTemplate{
							{
								Path:     "TestName",
								Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
							},
						},
					},
				},
				Status: heistv1alpha1.VaultBindingStatus{},
			}
			Expect(Test.K8sClient.Create(context.TODO(), binding)).To(Succeed())

			roleName := core.RoleName(fmt.Sprintf("managed.k8s.default.%s", saName))
			Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).WithTimeout(5 * time.Minute).Should(HavePolicies(
				core.PolicyName(fmt.Sprintf("managed.kv.%s.%s", secret.Namespace, secret.Name)),
			))
			Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToServiceAccounts(saName))
			Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToNamespaces("default"))

			Test.K8sEnv.VaultConfigSpec("default", saName).Should(testhelper.DeepEqual(&heistv1alpha1.VaultClientConfigSpec{
				Address:       "http://127.0.0.1:8200",
				Role:          "managed.k8s.default.some-sa-0",
				AuthMountPath: managed.KubernetesAuthPath,
				KvSecrets: []*heistv1alpha1.VaultKVSecretRef{
					{
						Name:       secret.Name,
						EnginePath: "managed/kv/default/engine-0",
						SecretPath: "secret-0",
						Capabilities: []heistv1alpha1.VaultBindingKVCapability{
							heistv1alpha1.VaultBindingKVCapabilityRead,
						},
					},
				},
				Templates: heistv1alpha1.VaultBindingAgentConfig{
					CertificateTemplates: nil,
					Templates: []heistv1alpha1.VaultBindingValueTemplate{
						{
							Path:     "TestName",
							Mode:     "0640",
							Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", secret.Name),
						},
					},
				},
			}))

			k8sRole := &v1.Role{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-client-config", saName),
					Namespace: "default",
				},
			}

			Eventually(func() error {
				return Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(k8sRole), k8sRole)
			}, 30*time.Second, 250*time.Millisecond).Should(Succeed())

			Expect(k8sRole.Rules).To(HaveLen(1))
			Expect(k8sRole.Rules[0].Verbs).To(Equal([]string{"get", "list", "watch"}))
			Expect(k8sRole.Rules[0].APIGroups).To(Equal([]string{"heist.youniqx.com"}))
			Expect(k8sRole.Rules[0].Resources).To(Equal([]string{"vaultclientconfigs"}))
			Expect(k8sRole.Rules[0].ResourceNames).To(BeNil())

			k8sRoleBinding := &v1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("%s-client-config", saName),
					Namespace: "default",
				},
			}

			Eventually(func() error {
				return Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(k8sRoleBinding), k8sRoleBinding)
			}, 30*time.Second, 250*time.Millisecond).Should(Succeed())

			Expect(k8sRoleBinding.RoleRef.APIGroup).To(Equal("rbac.authorization.k8s.io"))
			Expect(k8sRoleBinding.RoleRef.Kind).To(Equal("Role"))
			Expect(k8sRoleBinding.RoleRef.Name).To(Equal("some-sa-0-client-config"))

			Expect(k8sRoleBinding.Subjects).To(HaveLen(1))
			Expect(k8sRoleBinding.Subjects[0].Kind).To(Equal("ServiceAccount"))
			Expect(k8sRoleBinding.Subjects[0].APIGroup).To(BeEmpty())
			Expect(k8sRoleBinding.Subjects[0].Name).To(Equal("some-sa-0"))
			Expect(k8sRoleBinding.Subjects[0].Namespace).To(Equal("default"))
		})
	})

	It("Should correctly update VaultKubernetesAuthRole CRDs", func() {
		engine := &heistv1alpha1.VaultKVSecretEngine{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VaultKVSecretEngine",
				APIVersion: "heist.youniqx.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kv-engine",
				Namespace: "default",
			},
			Spec:   heistv1alpha1.VaultKVSecretEngineSpec{},
			Status: heistv1alpha1.VaultKVSecretEngineStatus{},
		}
		Expect(Test.K8sClient.Create(context.TODO(), engine)).To(Succeed())

		firstSecret := &heistv1alpha1.VaultKVSecret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VaultKVSecret",
				APIVersion: "heist.youniqx.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kv-engine-secret",
				Namespace: "default",
			},
			Spec: heistv1alpha1.VaultKVSecretSpec{
				Engine: engine.Name,
				Fields: map[string]*heistv1alpha1.VaultKVSecretField{
					"some-field": {
						CipherText: heistv1alpha1.EncryptedValue(Test.DefaultCipherText),
					},
				},
				DeleteProtection: false,
			},
			Status: heistv1alpha1.VaultKVSecretStatus{},
		}
		Expect(Test.K8sClient.Create(context.TODO(), firstSecret)).To(Succeed())
		Test.VaultEnv.KvSecret(engine, firstSecret).Should(HaveKvSecretFieldWithValue("some-field", "ASDF ASDF"))

		secondSecret := &heistv1alpha1.VaultKVSecret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VaultKVSecret",
				APIVersion: "heist.youniqx.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "second-kv-engine-secret",
				Namespace: "default",
			},
			Spec: heistv1alpha1.VaultKVSecretSpec{
				Engine: engine.Name,
				Fields: map[string]*heistv1alpha1.VaultKVSecretField{
					"another-field": {
						CipherText: heistv1alpha1.EncryptedValue(Test.DefaultCipherText),
					},
				},
				DeleteProtection: false,
			},
			Status: heistv1alpha1.VaultKVSecretStatus{},
		}
		Expect(Test.K8sClient.Create(context.TODO(), secondSecret)).To(Succeed())
		Test.VaultEnv.KvSecret(engine, secondSecret).Should(HaveKvSecretFieldWithValue("another-field", "ASDF ASDF"))

		Eventually(testhelper.HaveCondition(heistv1alpha1.Conditions.Types.Provisioned, metav1.ConditionTrue, heistv1alpha1.Conditions.Types.Provisioned, "Secret has been provisioned"))

		By("Creating an Auth Role based on a binding")
		firstBinding := &heistv1alpha1.VaultBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VaultBinding",
				APIVersion: "heist.youniqx.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "some-binding",
				Namespace: "default",
			},
			Spec: heistv1alpha1.VaultBindingSpec{
				Subject: heistv1alpha1.VaultBindingSubject{
					Name: "some-service-account",
				},
				KVSecrets: []heistv1alpha1.VaultBindingKV{
					{
						Name: firstSecret.Name,
					},
				},
				Agent: heistv1alpha1.VaultBindingAgentConfig{
					Templates: []heistv1alpha1.VaultBindingValueTemplate{
						{
							Path:     "some-field",
							Template: fmt.Sprintf("{{ kvSecret \"%s\" \"some-field\" }}", firstSecret.Name),
						},
					},
				},
			},
			Status: heistv1alpha1.VaultBindingStatus{},
		}
		Expect(Test.K8sClient.Create(context.TODO(), firstBinding)).To(Succeed())
		roleName := core.RoleName("managed.k8s.default.some-service-account")
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(HavePolicies(
			core.PolicyName(fmt.Sprintf("managed.kv.%s.%s", firstSecret.Namespace, firstSecret.Name)),
		))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToServiceAccounts("some-service-account"))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToNamespaces("default"))

		By("Updating the Auth Role if a second binding is added")
		secondBinding := &heistv1alpha1.VaultBinding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "VaultBinding",
				APIVersion: "heist.youniqx.com/v1alpha1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "another-binding",
				Namespace: "default",
			},
			Spec: heistv1alpha1.VaultBindingSpec{
				Subject: heistv1alpha1.VaultBindingSubject{
					Name: "some-service-account",
				},
				KVSecrets: []heistv1alpha1.VaultBindingKV{
					{
						Name: secondSecret.Name,
					},
				},
				Agent: heistv1alpha1.VaultBindingAgentConfig{
					Templates: []heistv1alpha1.VaultBindingValueTemplate{
						{
							Path:     "another-field",
							Template: fmt.Sprintf("{{ kvSecret \"%s\" \"another-field\" }}", secondSecret.Name),
						},
					},
				},
			},
			Status: heistv1alpha1.VaultBindingStatus{},
		}
		// Wait 1s, because creationTimestamp which is used to get the dominant binding, has only second precision
		time.Sleep(1 * time.Second)

		Expect(Test.K8sClient.Create(context.TODO(), secondBinding)).To(Succeed())
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(HavePolicies(
			core.PolicyName(fmt.Sprintf("managed.kv.%s.%s", firstSecret.Namespace, firstSecret.Name)),
		))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToServiceAccounts("some-service-account"))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToNamespaces("default"))

		By("Updating the Auth Role if a binding is deleted")
		Expect(Test.K8sClient.Delete(context.TODO(), firstBinding)).To(Succeed())
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).WithTimeout(5 * time.Minute).Should(HavePolicies(
			core.PolicyName(fmt.Sprintf("managed.kv.%s.%s", secondSecret.Namespace, secondSecret.Name)),
		))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToServiceAccounts("some-service-account"))
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).Should(BeBoundToNamespaces("default"))

		By("Deleting the Auth Role if all bindings are deleted")
		Expect(Test.K8sClient.Delete(context.TODO(), secondBinding)).To(Succeed())
		Test.VaultEnv.KubernetesAuthRole(managed.KubernetesAuth, roleName).WithTimeout(5 * time.Minute).Should(BeNil())
	}, 60)
})
