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

package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VaultKVSecretEngine Webhooks", func() {
	It("Should validate VaultKVSecretEngine fields", func() {
		By("Allowing valid crds", func() {
			engine := &VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-engine",
					Namespace: "default",
				},
				Spec: VaultKVSecretEngineSpec{
					MaxVersions:      0,
					DeleteProtection: false,
				},
				Status: VaultKVSecretEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).To(Succeed())
		})
		By("Allowing valid positive numbers for max versions", func() {
			engine := &VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "another-engine",
					Namespace: "default",
				},
				Spec: VaultKVSecretEngineSpec{
					MaxVersions:      10,
					DeleteProtection: false,
				},
				Status: VaultKVSecretEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).To(Succeed())
		})
		By("Preventing max versions from being set to a negative value", func() {
			engine := &VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yet-another-engine",
					Namespace: "default",
				},
				Spec: VaultKVSecretEngineSpec{
					MaxVersions:      -10,
					DeleteProtection: false,
				},
				Status: VaultKVSecretEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).ToNot(Succeed())
		})
		By("Allowing delete protection to be enabled", func() {
			engine := &VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yet-another-new-engine",
					Namespace: "default",
				},
				Spec: VaultKVSecretEngineSpec{
					MaxVersions:      0,
					DeleteProtection: true,
				},
				Status: VaultKVSecretEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).To(Succeed())
		})
		By("Preventing objects with delete protection to be deleted", func() {
			engine := &VaultKVSecretEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultKVSecretEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yet-another-old-engine",
					Namespace: "default",
				},
				Spec: VaultKVSecretEngineSpec{
					MaxVersions:      0,
					DeleteProtection: true,
				},
				Status: VaultKVSecretEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).To(Succeed())
			Eventually(func() error {
				result := &VaultKVSecretEngine{}
				return K8sClient.Get(ctx, client.ObjectKeyFromObject(engine), result)
			}).ShouldNot(HaveOccurred())
			Expect(K8sClient.Delete(ctx, engine)).NotTo(Succeed())

			engine.Spec.DeleteProtection = false
			Expect(K8sClient.Update(ctx, engine)).To(Succeed())
			Eventually(func() error {
				return K8sClient.Delete(ctx, engine)
			}).Should(Succeed())
		})
	}, 60)
})
