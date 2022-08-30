package injector

import (
	"strings"

	"github.com/mattbaird/jsonpatch"
	corev1 "k8s.io/api/core/v1"
)

func addVolumes(target, volumes []corev1.Volume, base string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(volumes))
	first := len(target) == 0
	var value interface{}
	for _, v := range volumes {
		value = v
		path := base
		if first {
			first = false
			value = []corev1.Volume{v}
		} else {
			path += "/-"
		}

		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}
	return result
}

func addVolumeMounts(target, mounts []corev1.VolumeMount, base string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(mounts))
	first := len(target) == 0
	var value interface{}
	for _, v := range mounts {
		value = v
		path := base
		if first {
			first = false
			value = []corev1.VolumeMount{v}
		} else {
			path += "/-"
		}

		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}
	return result
}

func addEnvs(target []corev1.EnvVar, vars []corev1.EnvVar, base string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(vars))
	first := len(target) == 0
	var value interface{}
	for _, v := range vars {
		value = v
		path := base
		if first {
			first = false
			value = []corev1.EnvVar{v}
		} else {
			path += "/-"
		}

		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}
	return result
}

func addContainers(target, containers []corev1.Container, base string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(containers))
	first := len(target) == 0
	var value interface{}
	for _, container := range containers {
		value = container
		path := base
		if first {
			first = false
			value = []corev1.Container{container}
		} else {
			path += "/-"
		}

		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}

	return result
}

func addInitContainers(target, containers []corev1.Container, base string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(containers))
	first := len(target) == 0
	var value interface{}
	for i := len(containers) - 1; i >= 0; i-- {
		container := containers[i]
		value = container
		path := base
		if first {
			first = false
			value = []corev1.Container{container}
		} else {
			path += "/0"
		}

		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      path,
			Value:     value,
		})
	}

	return result
}

func updateAnnotations(target, annotations map[string]string) []*jsonpatch.JsonPatchOperation {
	result := make([]*jsonpatch.JsonPatchOperation, 0, len(annotations))

	if len(target) == 0 {
		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      "/metadata/annotations",
			Value:     annotations,
		})

		return result
	}

	for key, value := range annotations {
		result = append(result, &jsonpatch.JsonPatchOperation{
			Operation: "add",
			Path:      "/metadata/annotations/" + EscapeJSONPointer(key),
			Value:     value,
		})
	}

	return result
}

func EscapeJSONPointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}
