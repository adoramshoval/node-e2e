// Reference: https://github.com/kubernetes/kubernetes/blob/7b3589151285716cd7b0a002bab9f73c32c286df/test/e2e/framework/pod/resource.go#L299
package pod

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateContainerSpec(containerName string, image string, mounts []v1.VolumeMount, ports []v1.ContainerPort, resources v1.ResourceRequirements, args ...string) v1.Container {
	if len(args) == 0 {
		args = []string{"pause"}
	}
	return v1.Container{
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

// CreatePod gets pod metadata and specification as arguments and creates a Pod struct. This is a variadic function as it accepts N number of containers.
// Returns: pointer (*) to v1.Pod (https://pkg.go.dev/k8s.io/api/core/v1#Pod)
func CreatePod(ns string, podName string, volumes []v1.Volume, containers ...v1.Container) *v1.Pod {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: *CreatePodSpec(volumes, containers...),
	}
	return pod
}

func CreatePodSpec(volumes []v1.Volume, containers ...v1.Container) *v1.PodSpec {
	// immediate := int64(0)
	var defaultTerminationGracePeriod int64 = 30
	return &v1.PodSpec{
		Containers:                    containers[:],
		Volumes:                       volumes,
		SecurityContext:               &v1.PodSecurityContext{},
		TerminationGracePeriodSeconds: &defaultTerminationGracePeriod,
	}
}
