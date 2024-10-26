package nodes_test

import (
	"context"
	"fmt"
	utils "node-e2e/utils/node"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNodesReadiness(t *testing.T) {
	var nodesList v1.NodeList

	feat := features.New("Nodes Readiness").
		WithLabel("type", "Nodes").
		Assess("All nodes can be listed", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			if err := c.Client().Resources(namespace).List(ctx, &nodesList, resources.WithTimeout(fetchTimeout)); err != nil {
				t.Fatal(err)
			}
			t.Logf("Got %v %s", len(nodesList.Items), resourceType)
			if len(nodesList.Items) == 0 {
				t.Fatalf("Expected >0 %s but got %v", resourceType, len(nodesList.Items))
			}

			return ctx
		}).
		Assess("All nodes are in ready state", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var notReady bool
			for _, node := range nodesList.Items {
				if condition, ok := utils.IsNodePerfectState(&node); !ok {
					t.Errorf("Node %v is not perfectly ready: %v", node.Name, condition)
					notReady = true
				}
			}
			if notReady {
				t.Fatal("Not all nodes are ready")
			}
			t.Log("All nodes are ready and in perfect condition")

			return ctx
		}).
		Assess("All nodes system components are latest version", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// True if there is a diff in one of the nodes
			var diff bool

			// Take the system info of the first node as a reference
			refSystemInfo := nodesList.Items[0].Status.NodeInfo

			// Compare the system info of each node with the reference
			for _, node := range nodesList.Items[1:] { // Start from the second node
				t.Logf("Comparing %s vs %s", nodesList.Items[0].Name, node.Name)

				systemInfo := node.Status.NodeInfo
				if diffList, ok := nodeSystemInfoDiff(&refSystemInfo, &systemInfo); ok {
					diff = true

					for _, d := range diffList {
						t.Log(d)
					}
				}
			}
			if diff {
				t.Fatal("Not all nodes have the same SystemInfo")
			}
			t.Log("All nodes have the same SystemInfo")

			return ctx
		}).Feature()
	testsEnvironment.Test(t, feat)
}

func nodeSystemInfoDiff(refNode *v1.NodeSystemInfo, node *v1.NodeSystemInfo) ([]string, bool) {
	var diffList []string
	var diff bool

	// Create a slice of field names to compare
	fieldNames := []string{
		"KernelVersion",
		"OSImage",
		"ContainerRuntimeVersion",
		"KubeletVersion",
		"OperatingSystem",
		"Architecture",
	}

	refNodeVal := reflect.ValueOf(refNode).Elem()
	nodeVal := reflect.ValueOf(node).Elem()

	for _, fieldName := range fieldNames {
		refField := refNodeVal.FieldByName(fieldName) // returns value in fieldName
		nodeField := nodeVal.FieldByName(fieldName)   // returns value in fieldName

		if refField.IsValid() && nodeField.IsValid() && refField.String() != nodeField.String() {
			diffList = append(diffList, fmt.Sprintf("%s differs: %s vs %s", fieldName, refField.String(), nodeField.String()))
			diff = true
		}
	}
	return diffList, diff
}
