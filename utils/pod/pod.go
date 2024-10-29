// Reference: https://github.com/kubernetes/kubernetes/blob/7b3589151285716cd7b0a002bab9f73c32c286df/test/e2e/framework/pod/resource.go#L299
package pod

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenDefaultContainer(containerName string, image string, mounts []v1.VolumeMount, ports []v1.ContainerPort, resources v1.ResourceRequirements, args ...string) *v1.Container {
	if len(args) == 0 {
		args = []string{"pause"}
	}
	return &v1.Container{
		Name:            containerName,
		Image:           image,
		Args:            args,
		VolumeMounts:    mounts,
		Ports:           ports,
		Resources:       resources,
		SecurityContext: &v1.SecurityContext{},
		ImagePullPolicy: v1.PullIfNotPresent,
	}
}

// This function is used to create default corev1.PodSpec for many pod controllers. The term default is used in order to indicate this is a basic form
// of corev1.PodSpec. A pointer is returned, and the returned object can be patched with relevant and needed changes.
func GenDefaultPodSpec(volumes []v1.Volume, containers ...v1.Container) *v1.PodSpec {
	var defaultTerminationGracePeriod int64 = 30
	return &v1.PodSpec{
		Containers:                    containers[:],
		Volumes:                       volumes,
		SecurityContext:               &v1.PodSecurityContext{},
		TerminationGracePeriodSeconds: &defaultTerminationGracePeriod,
	}
}

// CreatePod gets pod metadata and specification as arguments and creates a Pod struct. This is a variadic function as it accepts N number of containers.
// Returns: pointer (*) to v1.Pod (https://pkg.go.dev/k8s.io/api/core/v1#Pod)
func GenDefaultPod(ns string, podName string, volumes []v1.Volume, containers ...v1.Container) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: *GenDefaultPodSpec(volumes, containers...),
	}
	return pod
}
