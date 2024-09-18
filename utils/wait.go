package utils

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	DefaultTimeout      int64 = 60
	DefaultTickInterval int64 = 1
)

type TimeoutConfig struct {
	Timeout    time.Duration
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

func GenerateDefaultTimeout() *TimeoutConfig {

	var timeout time.Duration = time.Duration(DefaultTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return &TimeoutConfig{Timeout: timeout, Ctx: ctx, CancelFunc: cancel}
}

func (tc *TimeoutConfig) NewTimeoutContext(timeout time.Duration) {
	tc.Timeout = timeout
	tc.Ctx, tc.CancelFunc = context.WithTimeout(context.Background(), timeout)
}

func GetPodWithTimeout(tc *TimeoutConfig, clientset *kubernetes.Clientset, namespace string, podName string) (*v1.Pod, error) {

	if tc.Ctx == nil || tc.CancelFunc == nil && tc.Timeout != 0 {
		tc.NewTimeoutContext(tc.Timeout)
	} else {
		tc = GenerateDefaultTimeout()
	}

	pod, err := clientset.CoreV1().Pods(namespace).Get(tc.Ctx, podName, metav1.GetOptions{})

	// ticker := time.NewTicker(tc.TickInterval)

	defer tc.CancelFunc()

	//    for i := 0; i < tc.Timeout.Seconds() / tc.TickInterval.Seconds(); i++ {
	//        select {
	//            case <-ticker.C:
	//                pod, err := clientset.CoreV1().Pods(namespace).Get(tc.Ctx, podName, metav1.GetOptions{})
	//	}
	//    }
	if err != nil {
		Logger.WithFields(logrus.Fields{
			"namespace": namespace,
			"podName":   podName,
		}).Errorf("Failed to retrieve pod: %v", err)
		return nil, err
	}
	return pod, nil
}
