package vaulttransitengine

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	heistv1alpha1 "github.com/youniqx/heist/pkg/apis/heist.youniqx.com/v1alpha1"
	. "github.com/youniqx/heist/pkg/testhelper"
	. "github.com/youniqx/heist/pkg/vault/matchers"
	"github.com/youniqx/heist/pkg/vault/transit"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("VaultTransitKey Controller", func() {
	When("creating a VaultTransitKey whose engine does not exist", func() {
		var engine *heistv1alpha1.VaultTransitEngine
		var key *heistv1alpha1.VaultTransitKey

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultTransitEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-transit-engine",
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultTransitEngineSpec{},
				Status: heistv1alpha1.VaultTransitEngineStatus{},
			}

			key = &heistv1alpha1.VaultTransitKey{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitKey",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "key",
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultTransitKeySpec{
					Engine: engine.Name,
					Type:   transit.TypeAes256Gcm96,
				},
				Status: heistv1alpha1.VaultTransitKeyStatus{},
			}
		})

		AfterEach(func() {
			Test.K8sEnv.DeleteIfPresent(engine, key)
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
		})

		It("should go into the Waiting state", func() {
			Test.K8sEnv.Create(key)
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionFalse,
				heistv1alpha1.Conditions.Reasons.ErrorConfig,
				"Referenced TransitEngine not found",
			))
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())
		})

		It("should be provisioned once the engine is created", func() {
			Test.K8sEnv.Create(key)
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionFalse,
				heistv1alpha1.Conditions.Reasons.ErrorConfig,
				"Referenced TransitEngine not found",
			))
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())

			Test.K8sEnv.Create(engine)
			Test.K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitEngine(engine).ShouldNot(BeNil())
		})
	})

	When("deleting an engine containing VaultTransitKey", func() {
		var engine *heistv1alpha1.VaultTransitEngine
		var key *heistv1alpha1.VaultTransitKey

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultTransitEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-transit-engine",
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultTransitEngineSpec{},
				Status: heistv1alpha1.VaultTransitEngineStatus{},
			}

			key = &heistv1alpha1.VaultTransitKey{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitKey",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "key",
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultTransitKeySpec{
					Engine: engine.Name,
					Type:   transit.TypeAes256Gcm96,
				},
				Status: heistv1alpha1.VaultTransitKeyStatus{},
			}

			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())

			Test.K8sEnv.Create(engine)
			Test.K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))

			Test.K8sEnv.Create(key)
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitEngine(engine).ShouldNot(BeNil())
		})

		AfterEach(func() {
			Test.K8sEnv.DeleteIfPresent(engine, key)
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
		})

		It("should force the secret into the ErrorConfig state again", func() {
			Expect(Test.K8sClient.Delete(context.TODO(), engine)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionFalse,
				heistv1alpha1.Conditions.Reasons.ErrorConfig,
				"Referenced TransitEngine not found",
			))
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())
		})
	})

	When("updating a VaultTransitKey", func() {
		var engine, engine2 *heistv1alpha1.VaultTransitEngine
		var key *heistv1alpha1.VaultTransitKey

		BeforeEach(func() {
			engine = &heistv1alpha1.VaultTransitEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-transit-engine",
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultTransitEngineSpec{},
				Status: heistv1alpha1.VaultTransitEngineStatus{},
			}

			engine2 = &heistv1alpha1.VaultTransitEngine{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitEngine",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default-transit-engine-2",
					Namespace: "default",
				},
				Spec:   heistv1alpha1.VaultTransitEngineSpec{},
				Status: heistv1alpha1.VaultTransitEngineStatus{},
			}

			key = &heistv1alpha1.VaultTransitKey{
				TypeMeta: metav1.TypeMeta{
					Kind:       "VaultTransitKey",
					APIVersion: "heist.youniqx.com/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "key",
					Namespace: "default",
				},
				Spec: heistv1alpha1.VaultTransitKeySpec{
					Engine: engine.Name,
					Type:   transit.TypeAes256Gcm96,
				},
				Status: heistv1alpha1.VaultTransitKeyStatus{},
			}

			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())

			Test.K8sEnv.Create(engine)
			Test.K8sEnv.Object(engine).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))

			Test.K8sEnv.Create(engine2)
			Test.K8sEnv.Object(engine2).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"Engine has been provisioned",
			))

			Test.K8sEnv.Create(key)
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitEngine(engine).ShouldNot(BeNil())
		})

		AfterEach(func() {
			Test.K8sEnv.DeleteIfPresent(engine, engine2, key)
			Test.VaultEnv.TransitKey(engine, key).Should(BeNil())
			Test.VaultEnv.TransitEngine(engine).Should(BeNil())
			Test.VaultEnv.TransitEngine(engine2).Should(BeNil())
		})

		It("should recreate the key for changes to KeyType", func() {
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitKey(engine, key).Should(HaveKeyType(transit.TypeAes256Gcm96))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			key.Spec.Type = transit.TypeED25519
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitKey(engine, key).Should(HaveKeyType(transit.TypeED25519))
		})

		It("should recreate the key for changes to Engine", func() {
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitKey(engine, key).Should(HaveKeyType(transit.TypeAes256Gcm96))
			Test.VaultEnv.TransitKey(engine, key).ShouldNot(BeNil())
			Test.VaultEnv.TransitKey(engine2, key).Should(BeNil())

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			key.Spec.Engine = engine2.Name
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
		})

		It("should recreate the key for changes to Exportable", func() {
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitKey(engine, key).Should(HaveKeyType(transit.TypeAes256Gcm96))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			key.Spec.Exportable = true
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(key.Spec.Exportable).To(BeTrue())
		})

		It("should recreate the key for changes to AllowPlaintextBackup", func() {
			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))
			Test.VaultEnv.TransitKey(engine, key).Should(HaveKeyType(transit.TypeAes256Gcm96))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			key.Spec.AllowPlaintextBackup = true
			Expect(Test.K8sClient.Update(context.TODO(), key)).To(Succeed())
			Test.K8sEnv.Object(key).Should(HaveCondition(
				heistv1alpha1.Conditions.Types.Provisioned,
				metav1.ConditionTrue,
				heistv1alpha1.Conditions.Reasons.Provisioned,
				"TransitKey has been provisioned",
			))

			Expect(Test.K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(key), key)).To(Succeed())
			Expect(key.Spec.AllowPlaintextBackup).To(BeTrue())
		})
	})
})
