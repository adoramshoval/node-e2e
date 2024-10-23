package config

import (
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/client-go/util/homedir"
)

// Flags' name
const (
	flagSAName                   = "sa-name"
	flagSAToken                  = "sa-token"
	flagClusterName              = "cluster-name"
	flagClusterEndpoint          = "cluster-endpoint"
	flagCertificateAuthorityData = "certificate-authority-data"
	flagDirName                  = "dir-name"
)

const (
	defaultDirName     string = "testdata"
	defaultContextName string = "default-context"
)

// Variables to store data provided from flags
var (
	saName                         string
	saToken                        string
	clusterName                    string
	clusterEndpoint                string
	certificateAuthorityDataBase64 string
	dirName                        string
)

type AuthenticationAttr struct {
	clusterName              string
	clusterEndpoint          string
	certificateAuthorityData string
	namespace                string
	saName                   string
	saToken                  string
	dirName                  string
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

func init() {
	// Define flags
	flag.StringVar(&saName, flagSAName, "", "ServiceAccount name")
	flag.StringVar(&saToken, flagSAToken, "", "ServiceAccount token for authentication")
	flag.StringVar(&clusterName, flagClusterName, "", "Kubernetes cluster name")
	flag.StringVar(&clusterEndpoint, flagClusterEndpoint, "", "Kubernetes cluster API server endpoint")
	flag.StringVar(&certificateAuthorityDataBase64, flagCertificateAuthorityData, "", "Base64 decoded value as a string of the API certificate authority")
	flag.StringVar(&dirName, flagDirName, "", "Directory name where the KubeConfig file will be stored (e.g testdata). By default KubeConfig will be stored at the user's home directory")
}

func New() *AuthenticationAttr {
	return &AuthenticationAttr{
		dirName: defaultDirName,
	}
}

func NewKubeConfig(a *AuthenticationAttr) (string, error) {
	// Check if all required attributes are provided
	if a.saName == "" || a.saToken == "" || a.clusterEndpoint == "" || a.certificateAuthorityData == "" {
		return "", errors.New("saName, saToken, clusterEndpoint and certificateAuthorityData must be provided")
	}

	// Validate and fix CA data
	fixedCert, err := validateAndFixBase64(a.certificateAuthorityData)
	if err != nil {
		return "", fmt.Errorf("certificate authority data is invalid: %v", err)
	}
	a.WithCertificateAuthorityData(fixedCert)

	// Set default namespace if not already set
	if a.namespace == "" {
		a.WithNamespace(getNamespace(a.namespace))
	}

	kubeConfig := a.genKubeConfig()
	path, err := a.saveKubeConfig(kubeConfig)

	return path, err
}

func NewKubeConfigFromFlags(a *AuthenticationAttr) (string, error) {
	// Parse the flags
	flag.Parse()

	// Check if all required flags are provided
	if saToken == "" || saName == "" || clusterEndpoint == "" || certificateAuthorityDataBase64 == "" {
		return "", errors.New("sa-name, sa-token, cluster-endpoint and certificate-authority-data flags must be provided")
	}

	if a == nil {
		a = New()
	}
	a.WithSAName(saName)
	a.WithSAToken(saToken)
	a.WithClusterEndpoint(clusterEndpoint)
	a.WithCertificateAuthorityData(certificateAuthorityDataBase64)
	a.WithClusterName(clusterName)
	if dirName != "" {
		a.WithDirName(dirName)
	}

	return NewKubeConfig(a)
}

func (a *AuthenticationAttr) genKubeConfig() *KubeConfig {
	// Create the kubeconfig struct
	kubeConfig := KubeConfig{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters: []Cluster{
			{
				Name: getClusterName(a.clusterName),
				Cluster: ClusterInfo{
					Server:                   a.clusterEndpoint,
					CertificateAuthorityData: a.certificateAuthorityData,
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
					Namespace: a.namespace,
				},
			},
		},
		CurrentContext: defaultContextName,
	}
	return &kubeConfig
}

func (a *AuthenticationAttr) saveKubeConfig(kc *KubeConfig) (string, error) {
	// Convert kubeconfig to YAML
	kubeConfigYAML, err := yaml.Marshal(kc)
	if err != nil {
		return "", err
	}

	home := homedir.HomeDir()

	kubeconfigdir := filepath.Join(home, a.dirName, ".kube")
	kubeconfigpath := filepath.Join(kubeconfigdir, "config")

	err = os.MkdirAll(kubeconfigdir, 0o777)
	if err != nil {
		return "", fmt.Errorf("failed to create KubeConfig directory %s: %v", kubeconfigdir, err)
	}

	err = createFile(kubeconfigpath, kubeConfigYAML)
	if err != nil {
		return "", fmt.Errorf("failed to create KubeConfig config file %s: %v", kubeconfigpath, err)
	}

	return kubeconfigpath, nil
}

func (a *AuthenticationAttr) WithClusterName(cn string) *AuthenticationAttr {
	a.clusterName = cn
	return a
}

func (a *AuthenticationAttr) WithClusterEndpoint(ce string) *AuthenticationAttr {
	a.clusterEndpoint = ce
	return a
}

func (a *AuthenticationAttr) WithCertificateAuthorityData(cad string) *AuthenticationAttr {
	a.certificateAuthorityData = cad
	return a
}

func (a *AuthenticationAttr) WithSAName(san string) *AuthenticationAttr {
	a.saName = san
	return a
}

func (a *AuthenticationAttr) WithSAToken(sat string) *AuthenticationAttr {
	a.saToken = sat
	return a
}

func (a *AuthenticationAttr) WithDirName(dn string) *AuthenticationAttr {
	a.dirName = dn
	return a
}

func (a *AuthenticationAttr) WithNamespace(n string) *AuthenticationAttr {
	a.namespace = n
	return a
}

func (a *AuthenticationAttr) GetServiceAccountName() string {
	return a.saName
}

func (a *AuthenticationAttr) GetServiceAccountToken() string {
	return a.saToken
}

func (a *AuthenticationAttr) GetServiceAccountNamespace() string {
	return a.namespace
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

func validateAndFixBase64(encodedData string) (string, error) {
	// First, sanitize the input
	sanitizedData := sanitizeInput(encodedData)

	// Try to decode it
	decodedData, err := base64.StdEncoding.DecodeString(sanitizedData)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %v", err)
	}

	// If decoding succeeds, re-encode it properly
	return base64.StdEncoding.EncodeToString(decodedData), nil
}

// sanitizeInput removes unnecessary whitespace, newlines, and other formatting issues
func sanitizeInput(encodedData string) string {
	// Remove any spaces or newlines
	return strings.TrimSpace(encodedData)
}
