package injector

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/youniqx/heist/pkg/testhelper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Agent Injector", func() {
	When("When handling a pod event for an annotated pod", func() {
		BeforeEach(func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent": "true",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									DefaultMode: &[]int32{420}[0],
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "main-container",
							Image:   "centos",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"sleep", "1000"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
						},
					},
					ServiceAccountName:           "my-service-account",
					AutomountServiceAccountToken: &[]bool{true}[0],
				},
			}
			K8sEnv.Create(pod)
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()
		})

		It("When handling an event for a pod with the injection annotation and injection enabled", func() {
			expectedPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-1",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent": "true",
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					DNSPolicy:                     "ClusterFirst",
					ServiceAccountName:            "my-service-account",
					DeprecatedServiceAccount:      "my-service-account",
					AutomountServiceAccountToken:  &[]bool{true}[0],
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        nil,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
						SeccompProfile:      nil,
					},
					SchedulerName: "default-scheduler",
					Tolerations: []corev1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
					},
					Priority:           &[]int32{0}[0],
					EnableServiceLinks: &[]bool{true}[0],
					PreemptionPolicy:   &[]corev1.PreemptionPolicy{corev1.PreemptLowerPriority}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
						{
							Name: "heist-path-heist",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "heist-agent-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HEIST_AGENT_URL",
									Value: "http://localhost:13037",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
						{
							Name:  "heist-agent",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "serve"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase:    corev1.PodPending,
					QOSClass: corev1.PodQOSBurstable,
				},
			}

			fetchedPod := &corev1.Pod{}
			Expect(K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(expectedPod), fetchedPod)).To(Succeed())
			fetchedPod.ManagedFields = nil
			fetchedPod.CreationTimestamp = metav1.Time{}
			fetchedPod.ResourceVersion = ""
			fetchedPod.UID = ""

			Expect(fetchedPod).To(DeepEqual(expectedPod))
		})
	})

	When("When handling an event for a pod with the injection annotation, but injection disabled", func() {
		BeforeEach(func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-2",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/agent-enabled": "false",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									DefaultMode: &[]int32{420}[0],
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "main-container",
							Image:   "centos",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"sleep", "1000"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
						},
					},
					ServiceAccountName:           "my-service-account",
					AutomountServiceAccountToken: &[]bool{true}[0],
				},
			}
			K8sEnv.Create(pod)
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()
		})

		It("Should correctly inject the agent into the pod", func() {
			expectedPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-2",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/agent-enabled": "false",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					DNSPolicy:                     "ClusterFirst",
					ServiceAccountName:            "my-service-account",
					DeprecatedServiceAccount:      "my-service-account",
					AutomountServiceAccountToken:  &[]bool{true}[0],
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        nil,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
						SeccompProfile:      nil,
					},
					SchedulerName: "default-scheduler",
					Tolerations: []corev1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
					},
					Priority:           &[]int32{0}[0],
					EnableServiceLinks: &[]bool{true}[0],
					PreemptionPolicy:   &[]corev1.PreemptionPolicy{corev1.PreemptLowerPriority}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
					},
				},
				Status: corev1.PodStatus{
					Phase:    corev1.PodPending,
					QOSClass: corev1.PodQOSBestEffort,
				},
			}

			fetchedPod := &corev1.Pod{}
			Expect(K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(expectedPod), fetchedPod)).To(Succeed())
			fetchedPod.ManagedFields = nil
			fetchedPod.CreationTimestamp = metav1.Time{}
			fetchedPod.ResourceVersion = ""
			fetchedPod.UID = ""

			Expect(fetchedPod).To(DeepEqual(expectedPod))
		})
	})

	When("When handling an event for a pod which already has an agent injected", func() {
		BeforeEach(func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-3",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent": "true",
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:           "my-service-account",
					AutomountServiceAccountToken: &[]bool{true}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
						{
							Name: "heist-path-heist",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "heist-agent-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HEIST_AGENT_URL",
									Value: "http://localhost:13037",
								},
							},
							ImagePullPolicy: corev1.PullAlways,
						},
						{
							Name:  "heist-agent",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "serve"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
						},
					},
				},
			}
			K8sEnv.Create(pod)
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()
		})

		It("Should correctly inject the agent into the pod", func() {
			expectedPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-3",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent": "true",
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					DNSPolicy:                     "ClusterFirst",
					ServiceAccountName:            "my-service-account",
					DeprecatedServiceAccount:      "my-service-account",
					AutomountServiceAccountToken:  &[]bool{true}[0],
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        nil,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
						SeccompProfile:      nil,
					},
					SchedulerName: "default-scheduler",
					Tolerations: []corev1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
					},
					Priority:           &[]int32{0}[0],
					EnableServiceLinks: &[]bool{true}[0],
					PreemptionPolicy:   &[]corev1.PreemptionPolicy{corev1.PreemptLowerPriority}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
						{
							Name: "heist-path-heist",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "heist-agent-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HEIST_AGENT_URL",
									Value: "http://localhost:13037",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
						{
							Name:  "heist-agent",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "serve"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase:    corev1.PodPending,
					QOSClass: corev1.PodQOSBurstable,
				},
			}

			fetchedPod := &corev1.Pod{}
			Expect(K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(expectedPod), fetchedPod)).To(Succeed())
			fetchedPod.ManagedFields = nil
			fetchedPod.CreationTimestamp = metav1.Time{}
			fetchedPod.ResourceVersion = ""
			fetchedPod.UID = ""

			Expect(fetchedPod).To(DeepEqual(expectedPod))
		})
	})

	When("When handling a pod event for an annotated pod with preload enabled", func() {
		BeforeEach(func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-4",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent":  "true",
						"heist.youniqx.com/agent-preload": "true",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									DefaultMode: &[]int32{420}[0],
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "main-container",
							Image:   "centos",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"sleep", "1000"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
						},
					},
					ServiceAccountName:           "my-service-account",
					AutomountServiceAccountToken: &[]bool{true}[0],
				},
			}
			K8sEnv.Create(pod)
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()
		})

		It("should be able to inject the agent and init containers", func() {
			expectedPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-4",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent":  "true",
						"heist.youniqx.com/agent-preload": "true",
						"heist.youniqx.com/agent-status":  "injected",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					DNSPolicy:                     "ClusterFirst",
					ServiceAccountName:            "my-service-account",
					DeprecatedServiceAccount:      "my-service-account",
					AutomountServiceAccountToken:  &[]bool{true}[0],
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        nil,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
						SeccompProfile:      nil,
					},
					SchedulerName: "default-scheduler",
					Tolerations: []corev1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
					},
					Priority:           &[]int32{0}[0],
					EnableServiceLinks: &[]bool{true}[0],
					PreemptionPolicy:   &[]corev1.PreemptionPolicy{corev1.PreemptLowerPriority}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
						{
							Name: "heist-path-heist",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "heist-agent-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "heist-agent-preload",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "sync"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HEIST_AGENT_URL",
									Value: "http://localhost:13037",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
						{
							Name:  "heist-agent",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "serve"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase:    corev1.PodPending,
					QOSClass: corev1.PodQOSBurstable,
				},
			}

			fetchedPod := &corev1.Pod{}
			Expect(K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(expectedPod), fetchedPod)).To(Succeed())
			fetchedPod.ManagedFields = nil
			fetchedPod.CreationTimestamp = metav1.Time{}
			fetchedPod.ResourceVersion = ""
			fetchedPod.UID = ""

			Expect(fetchedPod).To(DeepEqual(expectedPod))
		})
	})

	When("When handling a pod event for an annotated pod with init containers and preload enabled", func() {
		BeforeEach(func() {
			pod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-4",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent":  "true",
						"heist.youniqx.com/agent-preload": "true",
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									DefaultMode: &[]int32{420}[0],
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:    "init-container",
							Image:   "centos",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"sleep", "10"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:    "main-container",
							Image:   "centos",
							Command: []string{"/bin/bash", "-c"},
							Args:    []string{"sleep", "1000"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
						},
					},
					ServiceAccountName:           "my-service-account",
					AutomountServiceAccountToken: &[]bool{true}[0],
				},
			}
			K8sEnv.Create(pod)
		})

		AfterEach(func() {
			K8sEnv.CleanupCreatedObject()
		})

		It("should be able to inject the agent and init containers", func() {
			expectedPod := &corev1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod-4",
					Namespace: "default",
					Annotations: map[string]string{
						"heist.youniqx.com/inject-agent":  "true",
						"heist.youniqx.com/agent-preload": "true",
						"heist.youniqx.com/agent-status":  "injected",
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:                 "Always",
					TerminationGracePeriodSeconds: &[]int64{30}[0],
					DNSPolicy:                     "ClusterFirst",
					ServiceAccountName:            "my-service-account",
					DeprecatedServiceAccount:      "my-service-account",
					AutomountServiceAccountToken:  &[]bool{true}[0],
					SecurityContext: &corev1.PodSecurityContext{
						SELinuxOptions:      nil,
						WindowsOptions:      nil,
						RunAsUser:           nil,
						RunAsGroup:          nil,
						RunAsNonRoot:        nil,
						SupplementalGroups:  nil,
						FSGroup:             nil,
						Sysctls:             nil,
						FSGroupChangePolicy: nil,
						SeccompProfile:      nil,
					},
					SchedulerName: "default-scheduler",
					Tolerations: []corev1.Toleration{
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							Value:             "",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{300}[0],
						},
					},
					Priority:           &[]int32{0}[0],
					EnableServiceLinks: &[]bool{true}[0],
					PreemptionPolicy:   &[]corev1.PreemptionPolicy{corev1.PreemptLowerPriority}[0],
					Volumes: []corev1.Volume{
						{
							Name: "my-service-account-token-9p89g",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName:  "my-service-account-token-9p89g",
									Items:       nil,
									DefaultMode: &[]int32{420}[0],
									Optional:    nil,
								},
							},
						},
						{
							Name: "heist-path-heist",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
						{
							Name: "heist-agent-cache",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: nil,
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "heist-agent-preload",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "sync"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
						{
							Name:      "init-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "10"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
					},
					Containers: []corev1.Container{
						{
							Name:      "main-container",
							Image:     "centos",
							Command:   []string{"/bin/bash", "-c"},
							Args:      []string{"sleep", "1000"},
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "my-service-account-token-9p89g",
									ReadOnly:  false,
									MountPath: "/var/run/secrets/kubernetes.io/serviceaccount",
								},
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "HEIST_AGENT_URL",
									Value: "http://localhost:13037",
								},
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
							ImagePullPolicy:          corev1.PullAlways,
						},
						{
							Name:  "heist-agent",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "serve"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "v1",
											FieldPath:  "metadata.namespace",
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(25, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(25, resource.Milli),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: *resource.NewScaledQuantity(50, resource.Mega),
									corev1.ResourceCPU:    *resource.NewScaledQuantity(50, resource.Milli),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "heist-path-heist",
									ReadOnly:  false,
									MountPath: "/heist",
								},
								{
									Name:      "heist-agent-cache",
									ReadOnly:  false,
									MountPath: "/.cache",
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/live",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/ready",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: 13037,
										},
										Scheme: "HTTP",
									},
								},
								TimeoutSeconds:   1,
								PeriodSeconds:    10,
								SuccessThreshold: 1,
								FailureThreshold: 3,
							},
							ImagePullPolicy: "IfNotPresent",
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: nil,
									Drop: []corev1.Capability{
										"ALL",
									},
								},
								Privileged:               &[]bool{false}[0],
								RunAsUser:                &[]int64{65532}[0],
								RunAsGroup:               &[]int64{65532}[0],
								RunAsNonRoot:             &[]bool{true}[0],
								ReadOnlyRootFilesystem:   &[]bool{false}[0],
								AllowPrivilegeEscalation: &[]bool{false}[0],
							},
							TerminationMessagePath:   "/dev/termination-log",
							TerminationMessagePolicy: "File",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase:    corev1.PodPending,
					QOSClass: corev1.PodQOSBurstable,
				},
			}

			fetchedPod := &corev1.Pod{}
			Expect(K8sClient.Get(context.TODO(), client.ObjectKeyFromObject(expectedPod), fetchedPod)).To(Succeed())
			fetchedPod.ManagedFields = nil
			fetchedPod.CreationTimestamp = metav1.Time{}
			fetchedPod.ResourceVersion = ""
			fetchedPod.UID = ""

			Expect(fetchedPod).To(DeepEqual(expectedPod))
		})
	})
})
