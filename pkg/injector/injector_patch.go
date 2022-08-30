package injector

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mattbaird/jsonpatch"
	"github.com/youniqx/heist/pkg/controllers/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/strings/slices"
)

const (
	agentContainerName   = "heist-agent"
	preloadContainerName = "heist-agent-preload"
	defaultMountPath     = "/heist"
	agentURLEnvVarName   = "HEIST_AGENT_URL"
	agentMemoryRequest   = 25
	agentCPURequest      = 25
	agentMemoryLimits    = 50
	agentCPULimits       = 50
	defaultAgentPort     = 13037
	agentCacheVolumeName = "heist-agent-cache"
)

type Injector struct {
	Pod            *corev1.Pod
	MountPaths     []string
	Config         *Config
	PreloadSecrets bool
}

func (h *Handler) NewInjector(pod *corev1.Pod) (injector *Injector, err error) {
	mountPaths := []string{defaultMountPath}
	if customizedPaths, _ := common.GetAnnotationValue(pod, AnnotationAgentMountPaths); customizedPaths != "" {
		mountPaths = strings.Split(customizedPaths, ",")
		if !slices.Contains(mountPaths, defaultMountPath) {
			mountPaths = append(mountPaths, defaultMountPath)
		}
	}

	mountPaths = filterMountPaths(mountPaths)

	preloadValue, _ := common.GetAnnotationValue(pod, AnnotationAgentPreload)
	return &Injector{
		Pod:            pod,
		MountPaths:     mountPaths,
		PreloadSecrets: preloadValue == "true",
		Config:         h.Config,
	}, nil
}

func filterMountPaths(paths []string) []string {
	pathMap := make(map[string]bool)

	for _, path := range paths {
		for other, enabled := range pathMap {
			if !enabled {
				continue
			}

			if subPath, _ := filepath.Rel(other, path); !strings.Contains(subPath, "..") {
				pathMap[path] = false
				break
			}

			if subPath, _ := filepath.Rel(path, other); !strings.Contains(subPath, "..") {
				pathMap[other] = false
			}
		}

		if _, exists := pathMap[path]; !exists {
			pathMap[path] = true
		}
	}

	result := make([]string, 0, len(pathMap))
	for path, enabled := range pathMap {
		if enabled {
			result = append(result, path)
		}
	}

	sort.Strings(result)

	return result
}

func (i *Injector) Patch() ([]*jsonpatch.JsonPatchOperation, error) {
	var patches []*jsonpatch.JsonPatchOperation

	annotationPatches := updateAnnotations(
		i.Pod.ObjectMeta.Annotations,
		map[string]string{
			AnnotationAgentStatus: AgentStatusInjected,
		},
	)
	patches = append(patches, annotationPatches...)

	volumes := make([]corev1.Volume, 0, len(i.MountPaths)+1)
	for _, path := range i.MountPaths {
		volumes = append(volumes, corev1.Volume{
			Name: createVolumeMountPath(path),
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: corev1.StorageMediumMemory,
				},
			},
		})
	}
	volumes = append(volumes, corev1.Volume{
		Name: agentCacheVolumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMediumMemory,
			},
		},
	})

	volumePatches := addVolumes(
		i.Pod.Spec.Volumes,
		volumes,
		"/spec/volumes",
	)
	patches = append(patches, volumePatches...)

	mounts := make([]corev1.VolumeMount, 0, len(i.MountPaths))
	for _, path := range i.MountPaths {
		mounts = append(mounts, corev1.VolumeMount{
			Name:      createVolumeMountPath(path),
			MountPath: path,
			ReadOnly:  false,
		})
	}

	agentURL, agentContainer := i.agentContainer(mounts)

	containerPatches := addContainers(
		i.Pod.Spec.Containers,
		[]corev1.Container{
			*agentContainer,
		},
		"/spec/containers",
	)
	patches = append(patches, containerPatches...)

	for i, container := range i.Pod.Spec.InitContainers {
		mountPatches := addVolumeMounts(
			container.VolumeMounts,
			mounts,
			fmt.Sprintf("/spec/initContainers/%d/volumeMounts", i),
		)
		patches = append(patches, mountPatches...)
	}

	for i, container := range i.Pod.Spec.Containers {
		mountPatches := addVolumeMounts(
			container.VolumeMounts,
			mounts,
			fmt.Sprintf("/spec/containers/%d/volumeMounts", i),
		)
		patches = append(patches, mountPatches...)

		envPatches := addEnvs(
			container.Env,
			[]corev1.EnvVar{
				{
					Name:  agentURLEnvVarName,
					Value: agentURL,
				},
			},
			fmt.Sprintf("/spec/containers/%d/env", i),
		)
		patches = append(patches, envPatches...)
	}

	if i.PreloadSecrets {
		initContainer := i.preloadInitContainer(mounts)

		containerPatches := addInitContainers(
			i.Pod.Spec.InitContainers,
			[]corev1.Container{
				*initContainer,
			},
			"/spec/initContainers",
		)
		patches = append(patches, containerPatches...)
	}

	return patches, nil
}

func createVolumeMountPath(path string) string {
	segment := strings.ReplaceAll(path, "/", "-")
	segment = strings.Trim(segment, "-")
	return fmt.Sprintf("heist-path-%s", segment)
}

func (i *Injector) usesPort(port int32) bool {
	for _, container := range i.Pod.Spec.Containers {
		for _, p := range container.Ports {
			if p.ContainerPort == port {
				return true
			}
		}
	}

	return false
}

func (i *Injector) agentContainer(mounts []corev1.VolumeMount) (string, *corev1.Container) {
	image := i.Config.AgentImage

	if annotationImage, _ := common.GetAnnotationValue(i.Pod, AnnotationAgentImage); annotationImage != "" {
		image = annotationImage
	}

	var agentPort int32 = defaultAgentPort
	for i.usesPort(agentPort) {
		agentPort++
	}

	agentURL := fmt.Sprintf("http://localhost:%d", agentPort)

	return agentURL, i.createAgentContainer(mounts, image, agentPort)
}

func (i *Injector) createAgentContainer(mounts []corev1.VolumeMount, image string, agentPort int32) *corev1.Container {
	return &corev1.Container{
		Name:  agentContainerName,
		Image: image,
		Args:  []string{"agent", fmt.Sprintf("--address=:%d", agentPort), "serve"},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: *resource.NewScaledQuantity(agentMemoryRequest, resource.Mega),
				corev1.ResourceCPU:    *resource.NewScaledQuantity(agentCPURequest, resource.Milli),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: *resource.NewScaledQuantity(agentMemoryLimits, resource.Mega),
				corev1.ResourceCPU:    *resource.NewScaledQuantity(agentCPULimits, resource.Milli),
			},
		},
		VolumeMounts: append(mounts, corev1.VolumeMount{
			Name:      agentCacheVolumeName,
			MountPath: "/.cache",
			ReadOnly:  false,
		}),
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
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/live",
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: agentPort,
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
						IntVal: agentPort,
					},
				},
			},
		},
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: i.agentSecurityContext(),
	}
}

func (i *Injector) preloadInitContainer(mounts []corev1.VolumeMount) *corev1.Container {
	image := i.Config.AgentImage
	if annotationImage, _ := common.GetAnnotationValue(i.Pod, AnnotationAgentImage); annotationImage != "" {
		image = annotationImage
	}
	var agentPort int32 = defaultAgentPort
	for i.usesPort(agentPort) {
		agentPort++
	}

	return i.createPreloadInitContainer(mounts, image, agentPort)
}

func (i *Injector) createPreloadInitContainer(mounts []corev1.VolumeMount, image string, agentPort int32) *corev1.Container {
	return &corev1.Container{
		Name:  preloadContainerName,
		Image: image,
		Args:  []string{"agent", fmt.Sprintf("--address=:%d", agentPort), "sync"},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: *resource.NewScaledQuantity(agentMemoryRequest, resource.Mega),
				corev1.ResourceCPU:    *resource.NewScaledQuantity(agentCPURequest, resource.Milli),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: *resource.NewScaledQuantity(agentMemoryLimits, resource.Mega),
				corev1.ResourceCPU:    *resource.NewScaledQuantity(agentCPULimits, resource.Milli),
			},
		},
		VolumeMounts: append(mounts, corev1.VolumeMount{
			Name:      agentCacheVolumeName,
			MountPath: "/.cache",
			ReadOnly:  false,
		}),
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
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: i.agentSecurityContext(),
	}
}

func (i *Injector) agentSecurityContext() *corev1.SecurityContext {
	var (
		privileged                     = false
		runAsUser                int64 = 65532
		runAsGroup               int64 = 65532
		runAsNonRoot                   = true
		readOnlyRootFilesystem         = false
		allowPrivilegeEscalation       = false
	)

	if i.Config.OpenShift {
		return &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{
					"ALL",
				},
			},
			Privileged:               &privileged,
			RunAsNonRoot:             &runAsNonRoot,
			ReadOnlyRootFilesystem:   &readOnlyRootFilesystem,
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		}
	}

	return &corev1.SecurityContext{
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
	}
}
