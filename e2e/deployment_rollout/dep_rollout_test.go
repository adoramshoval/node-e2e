package deployment_rollout

import (
	"context"
	"fmt"
	"testing"
	"time"

	"node-e2e/utils"
	"node-e2e/utils/deployment"
	selector "node-e2e/utils/label_selector"
	"node-e2e/utils/pod"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestDeploymentRollout(t *testing.T) {
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

	// Define Deployment itself and use that struct for later operations against the cluster
	dep := deployment.GenDefaultDeployment(workloadName, namespace, deploymentReplicas, *selector.GenMatchLabels(testLabels), testLabels, *ps)

	// Define feature
	feat := features.New("Deployment Creation and Rollout").
		WithLabel("type", "Deployment").
		Assess("Test Deployment resource can be created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			if err := c.Client().Resources(namespace).Create(ctx, dep); !apierrors.IsAlreadyExists(err) && err != nil {
				t.Fatal(err)
			}

			t.Logf("Deployment %s has been created successfully", dep.ObjectMeta.GetName())

			return ctx
		}).
		Assess("Deployment was able to deploy", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).DeploymentAvailable(workloadName, namespace),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}

			t.Logf("Deployment %s is available", dep.ObjectMeta.GetName())
			return ctx
		}).
		Assess("Deployment successful rollout", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Get all pods under the Deployment for later deletion verification
			var podList corev1.PodList

			if err := c.Client().Resources(namespace).List(ctx, &podList, resources.WithLabelSelector(getFirstlabel(testLabels))); err != nil {
				t.Fatal(err)
			}
			// Patch Deployment with kubectl.kubernetes.io/restartedAt annotation to trigger a rollout
			patchData := []byte(fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format(time.RFC3339)))
			if err := c.Client().Resources(namespace).Patch(ctx, dep, k8s.Patch{PatchType: types.MergePatchType, Data: patchData}); err != nil {
				t.Fatal(err)
			}
			t.Logf("Deployment, %s, was triggered for a rollout", dep.ObjectMeta.GetName())

			// Wait for all pods to get deleted
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourcesDeleted(&podList),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}
			t.Log("All pods were deleted")

			// Fetch new pods
			if err := c.Client().Resources(namespace).List(ctx, &podList, resources.WithLabelSelector(getFirstlabel(testLabels))); err != nil {
				t.Fatal(err)
			}

			for _, pod := range podList.Items {
				if err := wait.For(conditions.New(c.Client().Resources(namespace)).PodReady(&pod),
					wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
					wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
					t.Fatal(err)
				}
			}
			t.Logf("Deployment, %s, rolled out successfully", dep.ObjectMeta.GetName())

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Delete the Deployment itself
			if err := c.Client().Resources(namespace).Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}

			// Wait for it to get deleted
			if err := wait.For(conditions.New(c.Client().Resources(namespace)).ResourceDeleted(dep),
				wait.WithInterval(time.Duration(pollIntervalSeconds)*time.Second),
				wait.WithTimeout(time.Duration(pollTimeoutMinutes)*time.Minute)); err != nil {
				t.Fatal(err)
			}
			t.Logf("Deployment, %s, is being deleted", dep.ObjectMeta.GetName())

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

			t.Logf("Deployment, %s, deleted", dep.ObjectMeta.GetName())
			return ctx
		}).
		Feature()

	testsEnvironment.Test(t, feat)
}

func getFirstlabel(lables map[string]string) string {
	var label string
	for key, value := range lables {
		label = fmt.Sprintf("%s=%s", key, value)
		break
	}
	return label
}
