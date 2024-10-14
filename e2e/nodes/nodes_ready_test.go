package nodes_test

import (
	"context"
	"node-e2e/utils"
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNodesReadiness(t *testing.T) {
	var nodesList v1.NodeList

	feat := features.New("Nodes Readiness").
		WithLabel("type", "Nodes").
		Assess("All nodes can be listed", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {

			if err := c.Client().Resources(namespace).List(ctx, &nodesList); err != nil {
				t.Fatal(err)
			}
			t.Logf("Got %v %s", len(nodesList.Items), resources)
			if len(nodesList.Items) == 0 {
				t.Fatalf("Expected >0 %s but got %v", resources, len(nodesList.Items))
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
				systemInfo := node.Status.NodeInfo

				if attrDiff, ok := utils.SystemInfoDifference(&refSystemInfo, &systemInfo); !ok {
					t.Errorf("Node %s has different SystemInfo than the reference node %s", node.Name, nodesList.Items[0].Name)
					diff = true
					for _, d := range attrDiff {
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
