package escalation

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	saNameAnnot string = "kubernetes.io/service-account.name"
)

type sa struct {
	name      string
	namespace string
	token     string
}

func New(name, namespace, token string) *sa {
	return &sa{
		name:      name,
		namespace: namespace,
		token:     token,
	}
}

// This will create a new ServiceAccount in a given Namespace and will update the Config.Client with the appropriate new token.
// This function is intended to be used inside a test function to later switch between privileged and unprivileged accounts.
func NewServiceAccount(name, ns string) func(ctx context.Context, t *testing.T, c *envconf.Config) (*sa, error) {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) (*sa, error) {
		// Generate a default ServiceAccount
		var servacc *corev1.ServiceAccount = genDefaultServiceAccount(name, ns)

		// Create the ServiceAccount
		if err := c.Client().Resources(ns).Create(ctx, servacc); err != nil {
			return nil, err
		}

		// Find the SA's token
		token, err := FindToken(name, ns)(ctx, t, c)
		if err != nil {
			return nil, err
		}

		return New(name, ns, token), nil
	}
}

func NewFromExisting(name, ns string) func(ctx context.Context, t *testing.T, c *envconf.Config) (*sa, error) {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) (*sa, error) {
		// Try getting the ServiceAccount
		if err := c.Client().Resources(ns).Get(ctx, name, ns, &corev1.ServiceAccount{}); err != nil {
			return nil, err
		}

		// Find the SA's token
		token, err := FindToken(name, ns)(ctx, t, c)
		if err != nil {
			return nil, err
		}

		return New(name, ns, token), nil
	}
}

func SwitchAccount(new *sa) func(ctx context.Context, t *testing.T, c *envconf.Config) (old *sa) {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) (old *sa) {

		// Get current configured sa
		current := GetCurrent()(ctx, t, c)
		// Inject the new sa's token into the *rest.Config
		c.Client().RESTConfig().BearerToken = new.token
		c.Client().RESTConfig().BearerTokenFile = ""

		return current

	}
}

func FindToken(name, ns string) func(ctx context.Context, t *testing.T, c *envconf.Config) (string, error) {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) (string, error) {
		var token string

		// Start token search
		var secretList corev1.SecretList
		if err := c.Client().Resources(ns).List(ctx, &secretList); err != nil {
			return "", err
		}

		// Search for the SA's token
		for _, sec := range secretList.Items {
			if metav1.HasAnnotation(sec.ObjectMeta, saNameAnnot) {
				if sec.ObjectMeta.Annotations[saNameAnnot] == name && sec.Type == corev1.SecretTypeServiceAccountToken {
					if tokenData, err := decodeBase64(sec.StringData["token"]); err != nil {
						return "", err
					} else {
						token = tokenData
					}
				}
			}
		}

		if token == "" {
			return "", fmt.Errorf("could not find token for ServiceAccount %s in namespace %s", name, ns)
		}

		return token, nil
	}
}

func GetCurrent() func(ctx context.Context, t *testing.T, c *envconf.Config) *sa {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) *sa {
		username := c.Client().RESTConfig().Username
		if username == "" {
			username = "default"
		}
		return New(username, "", c.Client().RESTConfig().BearerToken)
	}
}

func (s *sa) GetToken() string {
	return s.token
}

func (s *sa) WithName(name string) *sa {
	s.name = name
	return s
}

func (s *sa) WithNamespace(ns string) *sa {
	s.namespace = ns
	return s
}

func (s *sa) WithToken(token string) *sa {
	s.token = token
	return s
}

// func CleanUp() {

// }

// This will create a new ClusterRole using a provided file path and a ClusterRoleBinding which will assign the newly created ClusterRole
// to the passed ServiceAccount
func (s *sa) AssignClusterRole(crPath string) func(ctx context.Context, t *testing.T, c *envconf.Config) error {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) error {
		if !filepath.IsAbs(crPath) {
			return fmt.Errorf("absulute path must be provided, got %s", crPath)
		}
		if !fileExists(crPath) {
			return fmt.Errorf("file %s does not exist", crPath)
		}

		// Extract the directory and the file name
		dir := filepath.Dir(crPath)
		baseFilePath := filepath.Base(crPath)

		// Use os.DirFS to create an fs.FS for the directory
		fsys := os.DirFS(dir)

		cr := &rbacv1.ClusterRole{}
		if err := decoder.DecodeFile(fsys, baseFilePath, cr); err != nil {
			return err
		}

		// Attemt to create the ClusterRole with the decoded value
		if err := c.Client().Resources(s.namespace).Create(ctx, cr); err != nil {
			return err
		}

		crb := genDefaultClusterRoleBinding(s.name, s.namespace, cr.ObjectMeta.GetName())
		// Attemt to create the ClusterRoleBinding
		if err := c.Client().Resources(s.namespace).Create(ctx, crb); err != nil {
			return err
		}

		return nil
	}
}

func genDefaultServiceAccount(name, ns string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func genDefaultClusterRoleBinding(name, ns, crName string) *rbacv1.ClusterRoleBinding {
	crb := &rbacv1.ClusterRoleBinding{
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      name,
				Namespace: ns,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     crName,
		},
	}
	return crb
}

func decodeBase64(data string) (string, error) {
	decodedData, err := base64.StdEncoding.DecodeString(data)
	return string(decodedData), err

}

// fileExists reports whether the named file or directory exists.
func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}
