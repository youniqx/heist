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

package agent_test

import (
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	"github.com/youniqx/heist/pkg/controllers"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/testhelper"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/testenv"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	K8sClient         client.Client
	TestEnv           *envtest.Environment
	VaultEnv          testenv.TestEnv
	AgentConfig       *rest.Config
	RootAPI           vault.API
	K8sEnv            testhelper.KubernetesTestHelper
	DefaultCipherText string
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agent Suite")
}

var _ = BeforeSuite(BeforeSuiteSetup(), 60)

func BeforeSuiteSetup() func() {
	return func() {
		logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

		By("bootstrapping kubernetes environment")
		TestEnv = &envtest.Environment{
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join("..", "..", "config", "webhook")},
			},
			CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd", "bases")},
		}

		cfg, err := TestEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		By("bootstrapping vault environment")
		VaultEnv, err = testenv.StartTestEnv(8700)
		Expect(err).NotTo(HaveOccurred())
		Expect(VaultEnv).NotTo(BeNil())

		By("setup Vault for Operator")
		err = operator.SetupOperator(&operator.SetupConfig{
			VaultURL:             VaultEnv.GetAddress(),
			VaultToken:           "root",
			PolicyName:           "heist",
			RoleName:             "heist",
			HeistNamespace:       "heist-system",
			HeistServiceAccount:  "heist",
			KubernetesHost:       "https://kubernetes.default.svc.cluster.local",
			KubernetesJWTIssuer:  "",
			KubernetesJWTCACert:  string(cfg.CAData),
			KubernetesJWTPemKeys: nil,
		})
		Expect(err).NotTo(HaveOccurred())

		RootAPI, err = VaultEnv.GetAPI()
		Expect(err).NotTo(HaveOccurred())
		Expect(RootAPI).NotTo(BeNil())

		token, err := VaultEnv.CreateToken("heist")
		Expect(err).NotTo(HaveOccurred())
		Expect(token).NotTo(BeEmpty())

		vaultAPI, err := vault.NewAPI().
			WithAddressFrom(core.Value(VaultEnv.GetAddress())).
			WithTokenFrom(core.Value(token)).
			Complete()
		Expect(err).NotTo(HaveOccurred())
		Expect(vaultAPI).NotTo(BeNil())

		Expect(heistv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
		Expect(admissionv1beta1.AddToScheme(scheme.Scheme)).To(Succeed())

		// +kubebuilder:scaffold:scheme

		By("start operator")
		webhookInstallOptions := &TestEnv.WebhookInstallOptions
		mgr, err := operator.Create().
			WithVaultAPI(vaultAPI).
			WithRestConfig(cfg).
			WithOptions(controllerruntime.Options{
				Scheme:             scheme.Scheme,
				Host:               webhookInstallOptions.LocalServingHost,
				Port:               webhookInstallOptions.LocalServingPort,
				CertDir:            webhookInstallOptions.LocalServingCertDir,
				LeaderElection:     false,
				MetricsBindAddress: "0",
			}).
			Register(heistv1alpha1.Component()).
			Register(controllers.Component(&controllers.Config{})).
			Complete()
		Expect(err).ToNot(HaveOccurred())
		Expect(mgr).ToNot(BeNil())

		K8sClient = mgr.GetClient()
		Expect(K8sClient).ToNot(BeNil())

		AgentConfig = mgr.GetConfig()

		K8sEnv = testhelper.New(cfg, K8sClient)

		DefaultCipherText, err = RootAPI.TransitEncrypt(managed.TransitEngine, managed.TransitKey, []byte("ASDF ASDF"))
		Expect(err).NotTo(HaveOccurred())
		Expect(DefaultCipherText).NotTo(BeEmpty())

		go func() {
			defer GinkgoRecover()
			err = mgr.Start(controllerruntime.SetupSignalHandler())
			Expect(err).ToNot(HaveOccurred(), "failed to run manager")
			gexec.KillAndWait(4 * time.Second)

			err := TestEnv.Stop()
			Expect(err).ToNot(HaveOccurred())
		}()
	}
}

var _ = AfterSuite(AfterSuiteTeardown())

func AfterSuiteTeardown() func() {
	return func() {
		By("tearing down the vault environment")
		vaultEnvError := VaultEnv.Stop()
		Expect(vaultEnvError).NotTo(HaveOccurred())
	}
}
