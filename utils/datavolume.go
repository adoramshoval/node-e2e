package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func GenerateDataVolumeTemplateSpec(storageRequests string) (*cdiv1beta1.DataVolumeSpec, error) {
	parsedStorageRequests, err := resource.ParseQuantity(storageRequests)
	if err != nil {
		return nil, fmt.Errorf("could not parse storage requests %v", storageRequests)
	}

	dataVolumeSpec := cdiv1beta1.DataVolumeSpec{
		Source: &cdiv1beta1.DataVolumeSource{
			HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
				URL: "https://example.com/disk-image.img",
			},
		},
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: parsedStorageRequests,
				},
			},
		},
		PriorityClassName: "high-priority",
		ContentType:       cdiv1beta1.DataVolumeKubeVirt, // Content type is "kubevirt"
		Preallocation:     boolPtr(true),
	}
	return &dataVolumeSpec, nil
}

func GenerateDataVolumeSourceHTTP()

// Helper function to create a bool pointer
func boolPtr(b bool) *bool {
	return &b
}

/*
Example DataVolumeTemplateSpec list for a VM
spec:
  dataVolumeTemplates:
    - apiVersion: cdi.kubevirt.io/v1beta1
      kind: DataVolume
      metadata:
        creationTimestamp: null
        name: rhel7-9-8bwxw5
      spec:
        source:
          pvc:
            name: rhel7-9-az-b
            namespace: openshift-virtualization-os-images
        storage:
          accessModes:
            - ReadWriteMany
          resources:
            requests:
              storage: 15Gi
          storageClassName: az-b
          volumeMode: Block
*/
