// Reference: https://github.com/kubernetes/kubernetes/blob/7b3589151285716cd7b0a002bab9f73c32c286df/test/e2e/framework/pod/resource.go#L299
package utils

import (
	"fmt"
	"reflect"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func parseQuantity(value string, resourceName string) *resource.Quantity {
	if value == "" {
		return nil // No value provided, omit the attribute
	}
	parsedQuantity, err := resource.ParseQuantity(value)
	if err != nil {
		// Logger from logger.go file under the same utils package (can be accessed within the same package)
		Logger.WithFields(logrus.Fields{
			"resourceName": resourceName,
			"value":        value,
		}).Errorf("Error parsing %s: %v", resourceName, err)
		return nil // Parsing failed, omit the attribute
	}
	return &parsedQuantity
}

func GenerateResourceRequirements(cpuRequests string, cpuLimits string, memoryRequests string, memoryLimits string) v1.ResourceRequirements {
	resourceRequirements := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}

	if cpuLimit := parseQuantity(cpuLimits, "CPU Limit"); cpuLimit != nil {
		resourceRequirements.Limits[v1.ResourceCPU] = *cpuLimit
	}

	if memoryLimit := parseQuantity(memoryLimits, "Memory Limit"); memoryLimit != nil {
		resourceRequirements.Limits[v1.ResourceMemory] = *memoryLimit
	}

	if cpuRequest := parseQuantity(cpuRequests, "CPU Request"); cpuRequest != nil {
		resourceRequirements.Requests[v1.ResourceCPU] = *cpuRequest
	}

	if memoryRequest := parseQuantity(memoryRequests, "Memory Request"); memoryRequest != nil {
		resourceRequirements.Requests[v1.ResourceMemory] = *memoryRequest
	}

	return resourceRequirements
}

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

// CreatePodSpec gets pod metadata and specification as arguments. This is a variadic function as it accepts N number of containers.
// Returns: pointer (*) to v1.Pod (https://pkg.go.dev/k8s.io/api/core/v1#Pod)
func CreatePodSpec(ns string, podName string, volumes []v1.Volume, containers ...v1.Container) *v1.Pod {
	immediate := int64(0)
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: ns,
		},
		Spec: v1.PodSpec{
			Containers:                    containers[:],
			Volumes:                       volumes,
			SecurityContext:               &v1.PodSecurityContext{},
			TerminationGracePeriodSeconds: &immediate,
		},
	}
	return pod
}

//func CreatePod() {
//
//}

func GetPod(tc *TimeoutConfig, clientset *kubernetes.Clientset, namespace string, podName string) (*v1.Pod, error) {
	result, err := tc.DoWithTimeout(
		func() (interface{}, bool) {
			pod, err := clientset.CoreV1().Pods(namespace).Get(tc.Ctx, podName, metav1.GetOptions{})

			if err != nil {
				FailedToRetrievePodErrorLog(namespace, podName, err)
				return nil, false
			}

			return pod, true
		})

	if err != nil {
		FailedToRetrievePodErrorLog(namespace, podName, err)
		return nil, err

	} else if pod, ok := result.(*v1.Pod); ok {
		SuccessfullyRetrievedPodLog(namespace, podName)
		return pod, nil

	} else {
		UnexpectedTypeErrorLog("v1.pod", Uncapitalize(reflect.TypeOf(pod).String()))
		return nil, fmt.Errorf("unexpected type %s", Uncapitalize(reflect.TypeOf(pod).String()))
	}

}

func WaitForPodRunningReady(tc *TimeoutConfig, clientset *kubernetes.Clientset, namespace string, podName string) error {
	_, err := tc.DoWithTimeout(
		func() (interface{}, bool) {
			// GetPod initialized with a different TimeoutConfig to ensure a different timeout context is used for fetching the pod
			pod, err := GetPod(&TimeoutConfig{}, clientset, namespace, podName)
			if err != nil {
				return nil, false
			}
			return nil, IsPodRunningReady(pod)
		})

	if err != nil {
		FailedToRetrievePodErrorLog(namespace, podName, err)
		return err
	}

	return nil

}

func IsPodRunningReady(pod *v1.Pod) bool {
	condition := GetPodCondition(pod.Status, v1.PodReady)
	return condition != nil && condition.Status == v1.ConditionTrue && pod.Status.Phase == v1.PodRunning
}

func GetPodCondition(status v1.PodStatus, condType v1.PodConditionType) *v1.PodCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}
