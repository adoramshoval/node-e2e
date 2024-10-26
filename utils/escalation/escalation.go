/*
This package is designed to set up test environments with the minimum required permissions, ensuring that tests run with the least privilege necessary.
By creating a dedicated ServiceAccount and assigning it a ClusterRole with only the specific permissions needed for each test, the security of the testing process is significantly enhanced.
This approach minimizes the risk of over-privileged access, ensuring that tests are isolated and only capable of performing authorized operations.
Achieving this, however, requires active cooperation from users, who must define appropriate ClusterRoles with carefully scoped, minimal permissions to meet the specific needs of each test.
*/

package escalation

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	saNameAnnot         string = "kubernetes.io/service-account.name"
	defaultSANamePrefix string = "test-sa"
	pollTimeoutMinutes  int64  = 1
	pollIntervalSeconds int64  = 3
)

type ServiceAccount struct {
	name      string
	namespace string
	token     string
}

func New(name, namespace, token string) *ServiceAccount {
	return &ServiceAccount{
		name:      name,
		namespace: namespace,
		token:     token,
	}
}

// This will create a new ServiceAccount in a given Namespace and will update the Config.Client with the appropriate new token.
// This function is intended to be used inside a test function to later switch between privileged and unprivileged accounts.
func NewServiceAccount(name, ns string) func(ctx context.Context, c *envconf.Config) (*ServiceAccount, error) {
	return func(ctx context.Context, c *envconf.Config) (*ServiceAccount, error) {

		// If the ServiceAccount happen to exist, return it
		s, err := NewFromExisting(name, ns)(ctx, c)
		if err == nil {
			return s, nil
		}

		// Generate a default ServiceAccount
		var servacc *corev1.ServiceAccount = genDefaultServiceAccount(name, ns)

		// Create the ServiceAccount
		if err := c.Client().Resources(ns).Create(ctx, servacc); !apierrors.IsAlreadyExists(err) && err != nil {
			return nil, err
		}

		// Wait until ServiceAccount has been created
		waitForObjectsCreation([]k8s.Object{servacc})(ctx, c)

		// Find the SA's token
		token, err := FindToken(name, ns)(ctx, c)
		if err != nil {
			return nil, err
		}

		return New(name, ns, token), nil
	}
}

// This will look for the ServiceAccount with the provided name and namespace and create sa struct from
// the SA's associated token.
func NewFromExisting(name, ns string) func(ctx context.Context, c *envconf.Config) (*ServiceAccount, error) {
	return func(ctx context.Context, c *envconf.Config) (*ServiceAccount, error) {
		// Try getting the ServiceAccount
		if err := c.Client().Resources(ns).Get(ctx, name, ns, &corev1.ServiceAccount{}); err != nil {
			return nil, err
		}

		// Find the SA's token
		token, err := FindToken(name, ns)(ctx, c)
		if err != nil {
			return nil, err
		}
		return New(name, ns, token), nil
	}
}

// This will set the provided ServiceAccount's token as the token to be used to authenticate against the cluster
// by injecting the token to the *rest.Config inside the *envconf.Config struct.
// This will also handle operation using the Resources struct as it holds a pointer to the *rest.Config as well.
func SwitchAccount(new *ServiceAccount) func(ctx context.Context, c *envconf.Config) (old *ServiceAccount, err error) {
	return func(ctx context.Context, c *envconf.Config) (old *ServiceAccount, err error) {

		// Get current configured sa
		current := GetCurrent()(ctx, c)
		if new.token == "" {
			return current, fmt.Errorf("can not switch to an empty token")
		}

		// Inject the new sa's token into the *rest.Config
		c.Client().RESTConfig().BearerToken = new.token
		c.Client().RESTConfig().BearerTokenFile = ""

		// Reinitialize the client to apply the new token
		client, err := klient.New(c.Client().RESTConfig())
		if err != nil {
			return current, fmt.Errorf("failed to initialize new client: %v", err)
		}

		// Update the config with the new client
		c.WithClient(client)

		return current, nil

	}
}

func FindToken(name, ns string) func(ctx context.Context, c *envconf.Config) (string, error) {
	return func(ctx context.Context, c *envconf.Config) (string, error) {
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
					if _, keyExist := sec.Data["token"]; keyExist {
						token = string(sec.Data["token"])
						break
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

// Since the *rest.Config does not store the ServiceAccount name, we are bound to use a replacement
// Returns a escalation.ServiceAccount with a random name
func GetCurrent() func(ctx context.Context, c *envconf.Config) *ServiceAccount {
	return func(ctx context.Context, c *envconf.Config) *ServiceAccount {
		username := c.Client().RESTConfig().Username
		if username == "" {
			username = envconf.RandomName(defaultSANamePrefix, 12)
		}
		return New(username, "", c.Client().RESTConfig().BearerToken)
	}
}

func (s *ServiceAccount) GetToken() string {
	return s.token
}

func (s *ServiceAccount) WithName(name string) *ServiceAccount {
	s.name = name
	return s
}

func (s *ServiceAccount) WithNamespace(ns string) *ServiceAccount {
	s.namespace = ns
	return s
}

func (s *ServiceAccount) WithToken(token string) *ServiceAccount {
	s.token = token
	return s
}

func (s *ServiceAccount) CleanUp(crPath string) func(ctx context.Context, c *envconf.Config) error {
	return func(ctx context.Context, c *envconf.Config) error {
		if !filepath.IsAbs(crPath) {
			return fmt.Errorf("absulute path must be provided, got %s", crPath)
		}
		if !fileExists(crPath) {
			return fmt.Errorf("file %s does not exist", crPath)
		}

		data, err := os.ReadFile(crPath)
		if err != nil {
			return err
		}

		cr := &rbacv1.ClusterRole{}
		if err := decoder.Decode(bytes.NewReader(data), cr); err != nil {
			return err
		}
		// Attemt to delete the ClusterRole with the decoded value
		if err := c.Client().Resources().Delete(ctx, cr); err != nil {
			return fmt.Errorf("error while deleting ClusterRole: %s: %v", cr.ObjectMeta.GetName(), err)
		}

		crb := genDefaultClusterRoleBinding(s.name, s.namespace, cr.ObjectMeta.GetName())
		// Attemt to delete the ClusterRoleBinding
		if err := c.Client().Resources(s.namespace).Delete(ctx, crb); err != nil {
			return fmt.Errorf("error while deleting ClusterRoleBinding: %s: %v", crb.ObjectMeta.GetName(), err)
		}

		sa := genDefaultServiceAccount(s.name, s.namespace)
		// Attempt to delete the ServiceAccount
		if err := c.Client().Resources(s.namespace).Delete(ctx, sa); err != nil {
			return fmt.Errorf("error while deleting ServiceAccount: %s: %v", s.name, err)
		}

		objList := []k8s.Object{cr, crb, sa}

		waitForObjectsDeletion(objList)(ctx, c)

		return nil
	}
}

// This will create a new ClusterRole using a provided file path and a ClusterRoleBinding which will assign the newly created ClusterRole
// to the passed ServiceAccount
func (s *ServiceAccount) AssignClusterRole(crPath string) func(ctx context.Context, c *envconf.Config) error {
	return func(ctx context.Context, c *envconf.Config) error {

		if !filepath.IsAbs(crPath) {
			return fmt.Errorf("absulute path must be provided, got %s", crPath)
		}
		if !fileExists(crPath) {
			return fmt.Errorf("file %s does not exist", crPath)
		}

		data, err := os.ReadFile(crPath)
		if err != nil {
			return err
		}

		cr := &rbacv1.ClusterRole{}
		if err := decoder.Decode(bytes.NewReader(data), cr); err != nil {
			return err
		}

		// Attemt to create the ClusterRole with the decoded value
		if err := c.Client().Resources().Create(ctx, cr); !apierrors.IsAlreadyExists(err) && err != nil {
			return err
		}

		crb := genDefaultClusterRoleBinding(s.name, s.namespace, cr.ObjectMeta.GetName())
		// Attemt to create the ClusterRoleBinding
		if err := c.Client().Resources(s.namespace).Create(ctx, crb); !apierrors.IsAlreadyExists(err) && err != nil {
			return err
		}

		waitForObjectsCreation([]k8s.Object{cr, crb})(ctx, c)

		return nil
	}
}

func waitForObjectsCreation(objList []k8s.Object) func(ctx context.Context, c *envconf.Config) error {
	return func(ctx context.Context, c *envconf.Config) error {
		for _, obj := range objList {
			if err := wait.For(conditions.New(c.Client().Resources()).ResourceMatch(obj, func(object k8s.Object) bool { return true }),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				return err
			}
		}
		return nil
	}
}

func waitForObjectsDeletion(objList []k8s.Object) func(ctx context.Context, c *envconf.Config) error {
	return func(ctx context.Context, c *envconf.Config) error {
		for _, obj := range objList {
			if err := wait.For(conditions.New(c.Client().Resources()).ResourceDeleted(obj),
				wait.WithTimeout(time.Minute*time.Duration(pollTimeoutMinutes)),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second)); err != nil {
				return err
			}
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
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-bind-%s", name, crName),
		},
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

// fileExists reports whether the named file or directory exists.
func fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}
