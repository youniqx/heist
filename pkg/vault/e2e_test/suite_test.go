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

package e2e_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/testenv"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	vaultEnv testenv.TestEnv
	vaultAPI vault.API
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Vault Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	var err error
	vaultEnv, err = testenv.StartTestEnv(8100)
	Expect(err).NotTo(HaveOccurred())
	Expect(vaultEnv).NotTo(BeNil())

	vaultAPI, err = vaultEnv.GetAPI()
	Expect(err).NotTo(HaveOccurred())
	Expect(vaultAPI).NotTo(BeNil())
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if vaultEnv != nil {
		Expect(vaultEnv.Stop()).To(Succeed())
	}
})
