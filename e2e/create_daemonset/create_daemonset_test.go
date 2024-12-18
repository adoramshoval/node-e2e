package createdaemonset_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"node-e2e/utils"
	selector "node-e2e/utils/label_selector"
	"node-e2e/utils/pod"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestDaemonSetCreation(t *testing.T) {
	var requestsCpu string = "100m"
	var requestsMemory string = "256Mi"

	// Define corev1.PodSpec
	ps := pod.GenDefaultPodSpec(nil, *pod.GenDefaultContainer(
		workloadName,
		testImage,
		nil,
		nil,
		corev1.ResourceRequirements{
			Requests: *utils.GenerateResourceList(requestsCpu, requestsMemory, "", ""),
			Limits:   *utils.GenerateResourceList(utils.GetCPULimitsFromRequests(requestsCpu), requestsMemory, "", ""),
		},
		"sleep", "100000000",
	))

	// Define DaemonSet itself and use that struct for later operations against the cluster
	ds := genDefaultDS(workloadName, namespace, selector.GenMatchLabels(testLabels), testLabels, *ps)
	// Set ServiceAccount name
	ds.Spec.Template.Spec.ServiceAccountName = saName
	// Set tolerations
	ds.Spec.Template.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "",       // Empty to match all taint keys
			Operator: "Exists", // "Exists" to match any taint value
			Effect:   "",       // Empty to match all taint effects (NoSchedule, PreferNoSchedule, NoExecute)
		},
	}

	feat := features.New("DaemonSet Creation and Interaction").
		WithLabel("type", "DaemonSet").
		Assess("Test DaemonSet resource can be created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			if err := c.Client().Resources(namespace).Create(ctx, ds); !apierrors.IsAlreadyExists(err) && err != nil {
				t.Fatal(err)
			}

			t.Logf("DaemonSet %s has been created successfully", ds.ObjectMeta.GetName())

			return ctx
		}).
		Assess("DaemonSet was able to deploy a pod on each available node", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).DaemonSetReady(ds),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}

			t.Logf("DaemonSet %s is ready and deployed a pod on each available node", ds.ObjectMeta.GetName())
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Delete the DaemonSet itself
			if err := c.Client().Resources(namespace).Delete(ctx, ds); err != nil {
				t.Fatal(err)
			}

			// Wait for it to get deleted
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(ds),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}

			// Fetch the underlying pod list
			var podList corev1.PodList
			if err := c.Client().Resources(namespace).List(ctx, &podList, resources.WithLabelSelector(getFirstlabel(testLabels))); err != nil {
				t.Fatal(err)
			}

			// Wait for all pods to get deleted
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourcesDeleted(&podList),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}

			t.Logf("DaemonSet %s deleted", ds.ObjectMeta.GetName())
			return ctx
		}).
		Feature()

	testsEnvironment.Test(t, feat)
}

// Helper function to generate DaemonSet configuration
func genDefaultDS(name, namespace string, labelsSelector *metav1.LabelSelector, podLabels map[string]string, podSpec corev1.PodSpec) *appsv1.DaemonSet {
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: labelsSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   name,
					Labels: podLabels,
				},
				Spec: podSpec,
			},
		},
	}
	return ds
}

func getFirstlabel(lables map[string]string) string {
	var label string
	for key, value := range lables {
		label = fmt.Sprintf("%s=%s", key, value)
		break
	}
	return label
}
