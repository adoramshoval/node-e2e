package utils

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    v1 "k8s.io/api/core/v1"
    resource "k8s.io/apimachinery/pkg/api/resource"
)

//func CreateContainerSpec () {
//
//}

// CreatePodSpec gets pod metadata and specification as arguments. This is a variadic function as it accepts N number of containers.
// Returns: pointer (*) to v1.Pod (https://pkg.go.dev/k8s.io/api/core/v1#Pod)
func CreatePodSpec(ns string, podName string, ns, podName string, volumes []v1.Volume, containers ...v1.Container) *v1.Pod {
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
//
//func GetPod() {
//
//}
