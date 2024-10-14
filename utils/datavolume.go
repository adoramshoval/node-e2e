package utils

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubev1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

type DataVolumeData struct {
	DVSource         *cdiv1beta1.DataVolumeSource
	PVAccessMode     string
	StorageRequests  string
	PVMode           string
	StorageClassName string
}

func GenerateDataVolumeTemplateSpec(dvname string, dvdata DataVolumeData) *kubev1.DataVolumeTemplateSpec {
	dvts := kubev1.DataVolumeTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: dvname,
		},
		Spec: *GenerateDataVolumeSpec(dvdata),
	}
	return &dvts
}

func GenerateDataVolume(dvname, ns string, dvdata DataVolumeData) *cdiv1beta1.DataVolume {
	dv := cdiv1beta1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dvname,
			Namespace: ns,
		},
		Spec: *GenerateDataVolumeSpec(dvdata),
	}
	return &dv
}

func GenerateDataVolumeSpec(dvdata DataVolumeData) *cdiv1beta1.DataVolumeSpec {
	dataVolumeSpec := cdiv1beta1.DataVolumeSpec{
		Source: dvdata.DVSource,
		PVC: &corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.PersistentVolumeAccessMode(dvdata.PVAccessMode),
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: *GenerateResourceList("", "", dvdata.StorageRequests, ""),
			},
			StorageClassName: &dvdata.StorageClassName,
			VolumeMode:       (*corev1.PersistentVolumeMode)(&dvdata.PVMode),
		},
		ContentType: cdiv1beta1.DataVolumeKubeVirt, // Content type is "kubevirt"
	}
	return &dataVolumeSpec
}

func GenerateDataVolumeSourceHTTP(sourceURL string) *cdiv1beta1.DataVolumeSource {
	dvsource := &cdiv1beta1.DataVolumeSource{
		HTTP: &cdiv1beta1.DataVolumeSourceHTTP{
			URL: sourceURL,
		},
	}
	return dvsource
}

func GenerateDataVolumeSourceRegistry(sourceURL, pullMethod *string) *cdiv1beta1.DataVolumeSource {
	dvsource := &cdiv1beta1.DataVolumeSource{
		Registry: &cdiv1beta1.DataVolumeSourceRegistry{
			URL:        sourceURL,
			PullMethod: (*cdiv1beta1.RegistryPullMethod)(pullMethod),
		},
	}
	return dvsource
}

func GenerateDataVolumeSourcePVC(namespace, pvcName string) *cdiv1beta1.DataVolumeSource {
	dvsource := &cdiv1beta1.DataVolumeSource{
		PVC: &cdiv1beta1.DataVolumeSourcePVC{
			Namespace: namespace,
			Name:      pvcName,
		},
	}
	return dvsource
}

func GenerateDataVolumeBlank() *cdiv1beta1.DataVolumeSource {
	dvsource := &cdiv1beta1.DataVolumeSource{
		Blank: &cdiv1beta1.DataVolumeBlankImage{},
	}
	return dvsource
}

func DataVolumeSourceSnapshot(namespace, volumeSnapshotName string) *cdiv1beta1.DataVolumeSource {
	dvsource := &cdiv1beta1.DataVolumeSource{
		Snapshot: &cdiv1beta1.DataVolumeSourceSnapshot{
			Namespace: namespace,
			Name:      volumeSnapshotName,
		},
	}
	return dvsource
}
