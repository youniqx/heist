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

package vaulttransitengine

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("VaultTransitEngine Controller", func() {
	It("Should provision engines correctly in Vault", func() {
		By("Being able to create a new engine")
		engine := &heistv1alpha1.VaultTransitEngine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-engine",
				Namespace: "default",
			},
			Spec:   heistv1alpha1.VaultTransitEngineSpec{},
			Status: heistv1alpha1.VaultTransitEngineStatus{},
		}
		Expect(Test.K8sClient.Create(context.TODO(), engine)).To(Succeed())
		Test.VaultEnv.TransitEngine(engine).Should(HavePath("managed/transit_engine/default/test-engine"))

		By("Being able to delete an existing engine")
		for i := 0; i < 3; i++ {
			err := Test.K8sClient.Delete(context.TODO(), engine)
			if err != nil {
				fmt.Printf("couldn't delete engine %s in test case 'Should provision engines correctly in Vault': %v", engine.GetName(), err)
			}
		}
		Test.K8sEnv.Object(engine).WithTimeout(1 * time.Minute).Should(BeNil())
		Test.VaultEnv.TransitEngine(engine).WithTimeout(time.Minute * 5).Should(BeNil())
	})
})
