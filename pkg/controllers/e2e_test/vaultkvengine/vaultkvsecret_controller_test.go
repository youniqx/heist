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

package vaultkvengine

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	. "github.com/youniqx/heist/pkg/testhelper"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VaultKVSecretEngine Controller", func() {
	It("Should provision engines correctly in Vault", func() {
		By("Being able to create a new engine")
		engine := &heistv1alpha1.VaultKVSecretEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-engine",
				Namespace: "default",
			},
			Spec:   heistv1alpha1.VaultKVSecretEngineSpec{},
			Status: heistv1alpha1.VaultKVSecretEngineStatus{},
		}
		Test.K8sEnv.Create(engine)
		Test.K8sEnv.Object(engine).Should(HaveCondition(
			heistv1alpha1.Conditions.Types.Provisioned,
			metav1.ConditionTrue,
			heistv1alpha1.Conditions.Reasons.Provisioned,
			"Engine has been provisioned",
		))
		Test.VaultEnv.KvEngine(engine).Should(HavePath("managed/kv/default/test-engine"))

		By("Being able to delete an existing engine")
		Expect(Test.K8sClient.Delete(context.TODO(), engine)).To(Succeed())
		Test.K8sEnv.Object(engine).Should(BeNil())
		Test.VaultEnv.KvEngine(engine).Should(BeNil())
	}, 60)
})
