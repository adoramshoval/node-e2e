package utils

import (
    "os"
    "github.com/sirupsen/logrus"
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
	Logger.Errorf("Could not find inClusterConfig: %v\n", err)	
        // Fallback to kubeconfig
        kubeconfigPath := os.Getenv("KUBECONFIG")
        if kubeconfigPath != "" {
            config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
        } else {
	    // Use the config from the home directory of the executing user
            config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}
    
	if err != nil {
	    Logger.Errorf("Could not find a valid kubeconfig: %v\n", err)
	    return nil, err
	}
    }

    // Create the Kubernetes clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
	Logger.Errorf("Could not create clientset for interfacing against the cluster: %v\n", err)
        return nil, err
    }
    
    Logger.WithFields(logrus.Fields{
    	"Host": config.Host,
	"APIPath": config.APIPath,
	"Username": config.Username,
	"ImpersonatedUser": config.Impersonate.UserName,
    }).Infof("Succussful authentication")
    return clientset, nil

}	

