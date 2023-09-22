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
)

var _ = Describe("VaultKVSecretEngine Webhooks", func() {
	It("Should validate VaultTransitEngine fields", func() {
		By("Allowing valid crds", func() {
			engine := &VaultTransitEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "some-engine",
					Namespace: "default",
				},
				Spec:   VaultTransitEngineSpec{},
				Status: VaultTransitEngineStatus{},
			}
			Expect(K8sClient.Create(ctx, engine)).To(Succeed())
		})
	})
})
