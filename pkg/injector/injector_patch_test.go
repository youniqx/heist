package injector

import (
	"reflect"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-test/deep"
	"github.com/mattbaird/jsonpatch"
	"github.com/youniqx/heist/pkg/operator"
	"github.com/youniqx/heist/pkg/vault"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestInjector_Patch(t *testing.T) {
	var (
		privileged                     = false
		runAsUser                int64 = 65532
		runAsGroup               int64 = 65532
		runAsNonRoot                   = true
		readOnlyRootFilesystem         = false
		allowPrivilegeEscalation       = false
	)

	type fields struct {
		Pod            *corev1.Pod
		MountPaths     []string
		Config         *Config
		PreloadSecrets bool
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*jsonpatch.JsonPatchOperation
		wantErr bool
	}{
		{
			name: "should be able to generate correct patches for default mount path",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to generate correct patches for default mount path in OpenShift",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					OpenShift:  true,
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to generate correct patches for pod which already uses the agent port",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
								Ports: []corev1.ContainerPort{
									{
										ContainerPort: 13037,
									},
								},
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13038", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
										IntVal: 13038,
									},
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13038,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13038",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to generate correct patches for default mount path and container with envs",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "SOME_VAR",
										Value: "SOME_VALUE",
									},
								},
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env/-",
					Value: corev1.EnvVar{
						Name:  "HEIST_AGENT_URL",
						Value: "http://localhost:13037",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to generate correct patches for default mount path and multi container pod with partial envs",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "SOME_VAR",
										Value: "SOME_VALUE",
									},
								},
							},
							{
								Name:  "another-app",
								Image: "another-app:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "ANOTHER_VAR",
										Value: "ANOTHER_VALUE",
									},
								},
							},
							{
								Name:  "yet-another-app",
								Image: "yet-another-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env/-",
					Value: corev1.EnvVar{
						Name:  "HEIST_AGENT_URL",
						Value: "http://localhost:13037",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/1/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/1/env/-",
					Value: corev1.EnvVar{
						Name:  "HEIST_AGENT_URL",
						Value: "http://localhost:13037",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/2/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/2/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to generate correct patches for multi container pod with partial envs and multiple mount paths",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "SOME_VAR",
										Value: "SOME_VALUE",
									},
								},
							},
							{
								Name:  "another-app",
								Image: "another-app:latest",
								Env: []corev1.EnvVar{
									{
										Name:  "ANOTHER_VAR",
										Value: "ANOTHER_VALUE",
									},
								},
							},
							{
								Name:  "yet-another-app",
								Image: "yet-another-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
					"/opt/mia/config",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-path-opt-mia-config",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
							},
							{
								Name:      "heist-path-opt-mia-config",
								ReadOnly:  false,
								MountPath: "/opt/mia/config",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts/-",
					Value: corev1.VolumeMount{
						Name:      "heist-path-opt-mia-config",
						MountPath: "/opt/mia/config",
						ReadOnly:  false,
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env/-",
					Value: corev1.EnvVar{
						Name:  "HEIST_AGENT_URL",
						Value: "http://localhost:13037",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/1/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/1/volumeMounts/-",
					Value: corev1.VolumeMount{
						Name:      "heist-path-opt-mia-config",
						MountPath: "/opt/mia/config",
						ReadOnly:  false,
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/1/env/-",
					Value: corev1.EnvVar{
						Name:  "HEIST_AGENT_URL",
						Value: "http://localhost:13037",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/2/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/2/volumeMounts/-",
					Value: corev1.VolumeMount{
						Name:      "heist-path-opt-mia-config",
						MountPath: "/opt/mia/config",
						ReadOnly:  false,
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/2/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to inject the preload init container if no init containers are present",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
				PreloadSecrets: true,
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/initContainers",
					Value: []corev1.Container{
						{
							Name:  "heist-agent-preload",
							Image: "youniqx/heist:latest",
							Args:  []string{"agent", "--address=:13037", "sync"},
							Env: []corev1.EnvVar{
								{
									Name: "AGENT_CLIENT_CONFIG_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "",
											FieldPath:  "spec.serviceAccountName",
										},
									},
								},
								{
									Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											APIVersion: "",
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
									Name:      "heist-path-vault-secrets",
									ReadOnly:  false,
									MountPath: "/vault/secrets",
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
								Privileged:               &privileged,
								RunAsUser:                &runAsUser,
								RunAsGroup:               &runAsGroup,
								RunAsNonRoot:             &runAsNonRoot,
								ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to inject the preload init container if an init container is already defined",
			fields: fields{
				Pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						InitContainers: []corev1.Container{
							{
								Name:  "some-init",
								Image: "some-init:latest",
							},
						},
						Containers: []corev1.Container{
							{
								Name:  "some-app",
								Image: "some-app:latest",
							},
						},
						ServiceAccountName: "some-account",
					},
				},
				MountPaths: []string{
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
				PreloadSecrets: true,
			},
			want: []*jsonpatch.JsonPatchOperation{
				{
					Operation: "add",
					Path:      "/metadata/annotations",
					Value: map[string]string{
						"heist.youniqx.com/agent-status": "injected",
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes",
					Value: []corev1.Volume{
						{
							Name: "heist-path-vault-secrets",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: "Memory",
								},
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/volumes/-",
					Value: corev1.Volume{
						Name: "heist-agent-cache",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{
								Medium: "Memory",
							},
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/-",
					Value: corev1.Container{
						Name:  "heist-agent",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "serve"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
								},
							},
						},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/ready",
									Port: intstr.IntOrString{
										Type:   intstr.Int,
										IntVal: 13037,
									},
								},
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/initContainers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/volumeMounts",
					Value: []corev1.VolumeMount{
						{
							Name:      "heist-path-vault-secrets",
							MountPath: "/vault/secrets",
							ReadOnly:  false,
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/containers/0/env",
					Value: []corev1.EnvVar{
						{
							Name:  "HEIST_AGENT_URL",
							Value: "http://localhost:13037",
						},
					},
				},
				{
					Operation: "add",
					Path:      "/spec/initContainers/0",
					Value: corev1.Container{
						Name:  "heist-agent-preload",
						Image: "youniqx/heist:latest",
						Args:  []string{"agent", "--address=:13037", "sync"},
						Env: []corev1.EnvVar{
							{
								Name: "AGENT_CLIENT_CONFIG_NAME",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
										FieldPath:  "spec.serviceAccountName",
									},
								},
							},
							{
								Name: "AGENT_CLIENT_CONFIG_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										APIVersion: "",
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
								Name:      "heist-path-vault-secrets",
								ReadOnly:  false,
								MountPath: "/vault/secrets",
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
							Privileged:               &privileged,
							RunAsUser:                &runAsUser,
							RunAsGroup:               &runAsGroup,
							RunAsNonRoot:             &runAsNonRoot,
							ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
							AllowPrivilegeEscalation: &allowPrivilegeEscalation,
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Injector{
				Pod:            tt.fields.Pod,
				MountPaths:     tt.fields.MountPaths,
				Config:         tt.fields.Config,
				PreloadSecrets: tt.fields.PreloadSecrets,
			}
			got, err := i.Patch()
			if (err != nil) != tt.wantErr {
				t.Errorf("Patch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("Patch() diff = %v", strings.Join(diff, "\n"))
			}
		})
	}
}

func TestHandler_NewInjector(t *testing.T) {
	type fields struct {
		Log           logr.Logger
		VaultAPI      vault.API
		K8sClient     client.Client
		Filter        operator.AnnotationFilter
		VaultAddress  string
		AuthMountPath string
		Config        *Config
	}
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantInjector *Injector
		wantErr      bool
	}{
		{
			name: "should be able to create injector with default mount path",
			fields: fields{
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			args: args{pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			}},
			wantInjector: &Injector{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: nil,
					},
				},
				MountPaths: []string{
					"/heist",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to create injector with additional mount path",
			fields: fields{
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			args: args{pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"heist.youniqx.com/agent-paths": "/heist/secrets,/opt/mia/config",
					},
				},
			}},
			wantInjector: &Injector{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/agent-paths": "/heist/secrets,/opt/mia/config",
						},
					},
				},
				MountPaths: []string{
					"/heist",
					"/opt/mia/config",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			wantErr: false,
		},
		{
			name: "should be able to create injector with different mount path",
			fields: fields{
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			args: args{pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"heist.youniqx.com/agent-paths": "/vault/secrets",
					},
				},
			}},
			wantInjector: &Injector{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"heist.youniqx.com/agent-paths": "/vault/secrets",
						},
					},
				},
				MountPaths: []string{
					"/heist",
					"/vault/secrets",
				},
				Config: &Config{
					AgentImage: "youniqx/heist:latest",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Log:           tt.fields.Log,
				VaultAPI:      tt.fields.VaultAPI,
				K8sClient:     tt.fields.K8sClient,
				Filter:        tt.fields.Filter,
				VaultAddress:  tt.fields.VaultAddress,
				AuthMountPath: tt.fields.AuthMountPath,
				Config:        tt.fields.Config,
			}
			gotInjector, err := h.NewInjector(tt.args.pod)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInjector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(gotInjector, tt.wantInjector); diff != nil {
				t.Errorf("NewInjector() diff = %v", strings.Join(diff, "\n"))
			}
		})
	}
}

func Test_filterMountPats(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "should filter out already recorded sub path",
			args: args{paths: []string{"/heist/secrets", "/heist"}},
			want: []string{"/heist"},
		},
		{
			name: "should filter out already recorded sub path when unrelated path has been recorded first",
			args: args{paths: []string{"/vault/secrets", "/heist/secrets", "/heist"}},
			want: []string{"/heist", "/vault/secrets"},
		},
		{
			name: "should filter out already recorded sub path when unrelated path has been recorded in the middle",
			args: args{paths: []string{"/heist/secrets", "/vault/secrets", "/heist"}},
			want: []string{"/heist", "/vault/secrets"},
		},
		{
			name: "should filter out already recorded sub path when unrelated path has been recorded after wards",
			args: args{paths: []string{"/heist/secrets", "/heist", "/vault/secrets"}},
			want: []string{"/heist", "/vault/secrets"},
		},
		{
			name: "should filter out path covered by other path",
			args: args{paths: []string{"/heist", "/heist/secrets"}},
			want: []string{"/heist"},
		},
		{
			name: "should filter out path covered by other path when unrelated path has been recorded first",
			args: args{paths: []string{"/vault/secrets", "/heist", "/heist/secrets"}},
			want: []string{"/heist", "/vault/secrets"},
		},
		{
			name: "should filter out path covered by other path when unrelated path has been recorded in the middle",
			args: args{paths: []string{"/heist", "/vault/secrets", "/heist/secrets"}},
			want: []string{"/heist", "/vault/secrets"},
		},
		{
			name: "should filter out path covered by other path when unrelated path has been recorded after wards",
			args: args{paths: []string{"/heist", "/heist/secrets", "/vault/secrets"}},
			want: []string{"/heist", "/vault/secrets"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterMountPaths(tt.args.paths); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterMountPaths() = %v, want %v", got, tt.want)
			}
		})
	}
}
