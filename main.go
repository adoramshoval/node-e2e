package main

import (
    "context"
    "fmt"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "node-e2e/utils"
)

const (
	namespace string = "core"
	podName string = "mypod"
)

func main() {
    // - Load kubeconfig
    // - Create the Kubernetes clientset
    clientset, err := utils.Authenticate()
    if err != nil {
        fmt.Println("Error creating clientset:", err)
        return
    }

    container1 := utils.CreateContainerSpec("container1", "quay.med.one:8443/openshift/ubi8/ubi", nil, nil, utils.GenerateResourceRequirements("250m", "1", "256Mi", "256Mi"), "sleep", "10000")
    container2 := utils.CreateContainerSpec("container2", "quay.med.one:8443/openshift/ubi8/ubi", nil, nil, utils.GenerateResourceRequirements("250m", "1", "256Mi", "256Mi"), "sleep", "10000")
    pod := utils.CreatePodSpec(namespace, podName, nil, container1, container2)

    result, err := clientset.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
    if err != nil {
        fmt.Println("Error creating Pod:", err)
    } else {
        fmt.Printf("Pod %s created in namespace %s\n", result.Name, result.Namespace)
    }

    tc := utils.GenerateDefaultTimeout()
    podSpec, err := utils.GetPodWithTimeout(tc, clientset, namespace, podName)
    fmt.Println(podSpec)

}
