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
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/youniqx/heist/pkg/managed"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	"github.com/youniqx/heist/pkg/vault/core"
	"github.com/youniqx/heist/pkg/vault/testenv"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	// +kubebuilder:scaffold:imports
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

func init() {
	logf.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))
}

var (
	K8sClient client.Client
	TestEnv   *envtest.Environment
	VaultEnv  testenv.TestEnv
	RootAPI   vault.API

	ctx    context.Context
	cancel context.CancelFunc
)

var defaultCipherText string

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	TestEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "config", "crd", "bases")},
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "..", "config", "webhook")},
		},
	}

	cfg, err := TestEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	VaultEnv, err = testenv.StartTestEnv(8300)
	Expect(err).NotTo(HaveOccurred())
	Expect(VaultEnv).NotTo(BeNil())

	RootAPI, err = VaultEnv.GetAPI()
	Expect(err).NotTo(HaveOccurred())
	Expect(RootAPI).NotTo(BeNil())

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

	scheme := runtime.NewScheme()
	Expect(AddToScheme(scheme)).To(Succeed())
	Expect(admissionv1beta1.AddToScheme(scheme)).To(Succeed())

	// +kubebuilder:scaffold:scheme

	webhookInstallOptions := &TestEnv.WebhookInstallOptions
	mgr, err := operator.Create().
		WithVaultAPI(vaultAPI).
		WithRestConfig(cfg).
		WithOptions(ctrl.Options{
			Scheme: scheme,
			WebhookServer: webhook.NewServer(webhook.Options{
				Host:    webhookInstallOptions.LocalServingHost,
				Port:    webhookInstallOptions.LocalServingPort,
				CertDir: webhookInstallOptions.LocalServingCertDir,
			}),
			LeaderElection: false,
			Metrics: metricsserver.Options{
				BindAddress: "0",
			},
		}).
		Register(Component()).
		Complete()
	Expect(err).ToNot(HaveOccurred())
	Expect(mgr).ToNot(BeNil())

	defaultCipherText, err = RootAPI.TransitEncrypt(managed.TransitEngine, managed.TransitKey, []byte("ASDF ASDF"))
	Expect(err).NotTo(HaveOccurred())
	Expect(defaultCipherText).NotTo(BeEmpty())

	go func() {
		err = mgr.Start(ctx)
		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		conn.Close()
		return nil
	}).Should(Succeed())

	K8sClient = mgr.GetClient()
	Expect(K8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	testEnvError := TestEnv.Stop()
	vaultEnvError := VaultEnv.Stop()
	Expect(testEnvError).NotTo(HaveOccurred())
	Expect(vaultEnvError).NotTo(HaveOccurred())
})
