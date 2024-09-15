package utils

import (
    "os"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

func Authenticate() (*kubernetes.Clientset, error) {
    var config *rest.Config
    var err error

    // Attempt to load in-cluster config - 
    // InClusterConfig returns a config object which uses the service account kubernetes gives to pods. 
    // It's intended for clients that expect to be running inside a pod running on kubernetes.
    config, err = rest.InClusterConfig()
    if err != nil {
        // Fallback to kubeconfig
        kubeconfigPath := os.Getenv("KUBECONFIG")
        if kubeconfigPath != "" {
            config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
        } else {
	    // Use the config from the home directory of the executing user
            config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}
    
	if err != nil {
	    return nil, err
	}
    }

    // Create the Kubernetes clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, err
    }
    
    return clientset, nil

}	

