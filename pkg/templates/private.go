package templates

import (
	corev1 "k8s.io/api/core/v1"
)

func convertEnvironentToEnvVar(environment map[string]string) []corev1.EnvVar {
	env := []corev1.EnvVar{}

	for key, value := range environment {
		env = append(env, corev1.EnvVar{
			Name:  key,
			Value: value,
		})
	}

	return env
}
