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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/vault/kvengine"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	"github.com/youniqx/heist/pkg/vault/policy"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ = Describe("PolicyAPI", func() {
	It("Should be able to manage policies", func() {
		engine := &kvengine.KvEngine{
			Path: "managed/engine",
			Config: &kvengine.Config{
				MaxVersions:        10,
				CasRequired:        true,
				DeleteVersionAfter: "0s",
			},
		}
		Expect(vaultAPI.UpdateKvEngine(engine)).To(Succeed())
		vaultEnv.KvEngine(engine).Should(HavePath("managed/engine"))
		vaultEnv.KvEngine(engine).Should(HaveConfig(&kvengine.Config{
			MaxVersions:        10,
			CasRequired:        true,
			DeleteVersionAfter: "0s",
		}))

		By("Creating a new one")
		newPolicy := &policy.Policy{
			Name: "managed.some-namespace.some-secret",
			Rules: []*policy.Rule{
				{
					Path: "/some/path/some-secret",
					Capabilities: []policy.Capability{
						policy.ReadCapability,
					},
				},
			},
		}
		Expect(vaultAPI.UpdatePolicy(newPolicy)).To(Succeed())
		vaultEnv.Policy(newPolicy).Should(HaveName("managed.some-namespace.some-secret"))
		vaultEnv.Policy(newPolicy).Should(HaveRules(
			&policy.Rule{
				Path: "/some/path/some-secret",
				Capabilities: []policy.Capability{
					policy.ReadCapability,
				},
			},
		))

		By("Not making an update if nothing has changed")
		Expect(vaultAPI.UpdatePolicy(newPolicy)).To(Succeed())
		vaultEnv.Policy(newPolicy).Should(HaveName("managed.some-namespace.some-secret"))
		vaultEnv.Policy(newPolicy).Should(HaveRules(
			&policy.Rule{
				Path: "/some/path/some-secret",
				Capabilities: []policy.Capability{
					policy.ReadCapability,
				},
			},
		))

		By("Updating it if something changed")
		newPolicy.Rules[0].Capabilities = append(newPolicy.Rules[0].Capabilities, policy.CreateCapability)
		Expect(vaultAPI.UpdatePolicy(newPolicy)).To(Succeed())
		vaultEnv.Policy(newPolicy).Should(HaveName("managed.some-namespace.some-secret"))
		vaultEnv.Policy(newPolicy).Should(HaveRules(
			&policy.Rule{
				Path: "/some/path/some-secret",
				Capabilities: []policy.Capability{
					policy.ReadCapability,
					policy.CreateCapability,
				},
			},
		))

		By("Deleting it")
		Expect(vaultAPI.DeletePolicy(newPolicy)).To(Succeed())
		vaultEnv.Policy(newPolicy).Should(BeNil())

		By("Not throwing an error when trying to delete a non existent one")
		Expect(vaultAPI.DeletePolicy(newPolicy)).To(Succeed())
		vaultEnv.Policy(newPolicy).Should(BeNil())
	})
})
