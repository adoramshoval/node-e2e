package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/util/homedir"
)

const (
	flagSAName          = "sa-name"
	flagSAToken         = "sa-token"
	flagClusterName     = "cluster-name"
	flagClusterEndpoint = "cluster-endpoint"
	flagDirName         = "dir-name"
)

const (
	defaultDirName     string = "testdata"
	defaultContextName string = "default-context"
)

var (
	saName          string
	saToken         string
	clusterName     string
	clusterEndpoint string
	dirName         string
)

type authenticationAttr struct {
	clusterName     string
	clusterEndpoint string
	namespace       string
	saName          string
	saToken         string
	dirName         string
}

type ClusterInfo struct {
	Server                   string `yaml:"server"`
	CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
}

type Cluster struct {
	Name    string      `yaml:"name"`
	Cluster ClusterInfo `yaml:"cluster"`
}

type UserInfo struct {
	Token string `yaml:"token"`
}

type User struct {
	Name string   `yaml:"name"`
	User UserInfo `yaml:"user"`
}

type ContextInfo struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Context struct {
	Name    string      `yaml:"name"`
	Context ContextInfo `yaml:"context"`
}

type KubeConfig struct {
	APIVersion     string    `yaml:"apiVersion"`
	Kind           string    `yaml:"kind"`
	Clusters       []Cluster `yaml:"clusters"`
	Users          []User    `yaml:"users"`
	Contexts       []Context `yaml:"contexts"`
	CurrentContext string    `yaml:"current-context"`
}

func New() *authenticationAttr {
	return &authenticationAttr{
		dirName: defaultDirName,
	}
}

func init() {
	// Define flags
	flag.StringVar(&saName, flagSAName, "", "ServiceAccount name")
	flag.StringVar(&saToken, flagSAToken, "", "ServiceAccount token for authentication")
	flag.StringVar(&clusterName, flagClusterName, "", "Kubernetes cluster name")
	flag.StringVar(&clusterEndpoint, flagClusterEndpoint, "", "Kubernetes cluster API server endpoint")
	flag.StringVar(&dirName, flagDirName, "", "Directory name where the KubeConfig file will be stored (e.g testdata). By default KubeConfig will be stored at the user's home directory")
}

func NewKubeConfig(a *authenticationAttr) error {
	// Check if all required attributes are provided
	if a.saName == "" || a.saToken == "" || a.clusterEndpoint == "" {
		return errors.New("saName, saToken and clusterEndpoint must be provided")
	}

	kubeConfig := a.genKubeConfig()
	err := a.saveKubeConfig(kubeConfig)

	return err
}

func NewKubeConfigFromFlags(a *authenticationAttr) error {
	// Parse the flags
	flag.Parse()

	// Check if all required flags are provided
	if saToken == "" || saName == "" || clusterEndpoint == "" {
		return errors.New("sa-name, sa-token and cluster-endpoint flags must be provided")
	}

	if a == nil {
		a = New()
	}
	a.WithSAName(saName)
	a.WithSAToken(saToken)
	a.WithClusterEndpoint(clusterEndpoint)
	a.WithClusterName(clusterName)
	if dirName != "" {
		a.WithDirName(dirName)
	}

	return NewKubeConfig(a)
}

func (a *authenticationAttr) genKubeConfig() *KubeConfig {
	// Create the kubeconfig struct
	kubeConfig := KubeConfig{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []Cluster{
			{
				Name: getClusterName(a.clusterName),
				Cluster: ClusterInfo{
					Server: a.clusterEndpoint,
				},
			},
		},
		Users: []User{
			{
				Name: a.saName,
				User: UserInfo{
					Token: a.saToken,
				},
			},
		},
		Contexts: []Context{
			{
				Name: defaultContextName,
				Context: ContextInfo{
					Cluster:   getClusterName(a.clusterName),
					User:      a.saName,
					Namespace: getNamespace(a.namespace),
				},
			},
		},
		CurrentContext: defaultContextName,
	}
	return &kubeConfig
}

func (a *authenticationAttr) saveKubeConfig(kc *KubeConfig) error {
	// Convert kubeconfig to YAML
	kubeConfigYAML, err := yaml.Marshal(kc)
	if err != nil {
		return err
	}

	home := homedir.HomeDir()

	kubeconfigdir := filepath.Join(home, a.dirName, ".kube")
	kubeconfigpath := filepath.Join(kubeconfigdir, "config")

	err = os.MkdirAll(kubeconfigdir, 0o777)
	if err != nil {
		return fmt.Errorf("failed to create KubeConfig directory %s: %v", kubeconfigdir, err)
	}

	err = createFile(kubeconfigpath, kubeConfigYAML)
	if err != nil {
		return fmt.Errorf("failed to create KubeConfig config file %s: %v", kubeconfigpath, err)
	}

	return nil
}

func (a *authenticationAttr) WithClusterName(cn string) *authenticationAttr {
	a.clusterName = cn
	return a
}

func (a *authenticationAttr) WithClusterEndpoint(ce string) *authenticationAttr {
	a.clusterEndpoint = ce
	return a
}

func (a *authenticationAttr) WithSAName(san string) *authenticationAttr {
	a.saName = san
	return a
}

func (a *authenticationAttr) WithSAToken(sat string) *authenticationAttr {
	a.saToken = sat
	return a
}

func (a *authenticationAttr) WithDirName(dn string) *authenticationAttr {
	a.dirName = dn
	return a
}

func (a *authenticationAttr) WithNamespace(n string) *authenticationAttr {
	a.namespace = n
	return a
}

func getNamespace(n string) string {
	if n == "" {
		return "default"
	}
	return n
}

func getClusterName(cn string) string {
	if cn == "" {
		return "default"
	}
	return cn
}

func createFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o644)
}
