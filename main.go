package main

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    resource "k8s.io/apimachinery/pkg/api/resource"
    "node-e2e/utils"
)

const (
	namespace string = "core"
)

func main() {
    // - Load kubeconfig
    // - Create the Kubernetes clientset
    clientset, err := utils.Authenticate()
    if err != nil {
        fmt.Println("Error creating clientset:", err)
        return
    }

    // Define the Pod
    pod := &corev1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: "mypod",
            Namespace: namespace,
        },
        Spec: corev1.PodSpec{
            Containers: []corev1.Container{
                {
                    Name:  "mycontainer",
                    Image: "quay.med.one:8443/openshift/rhel8/redis-6@sha256:7d4cb67a6416cb43d4b84663e9ede302f25df149febec132134dea5302c7c692",
                    Ports: []corev1.ContainerPort{
                        {
                            ContainerPort: 80,
                        },
                    },
		    Resources: corev1.ResourceRequirements{
  		        Limits: corev1.ResourceList{
			    corev1.ResourceCPU: resource.MustParse("1"),
			    corev1.ResourceMemory: resource.MustParse("256Mi"),
  		        },
		        Requests: corev1.ResourceList{
			    corev1.ResourceCPU: resource.MustParse("250m"),
                            corev1.ResourceMemory: resource.MustParse("256Mi"),
		        },
		    },
                },
            },
        },
    }

    // Create the Pod
    result, err := clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
    if err != nil {
        fmt.Println("Error creating Pod:", err)
    } else {
        fmt.Printf("Pod %s created in namespace %s\n", result.Name, result.Namespace)
    }

}
