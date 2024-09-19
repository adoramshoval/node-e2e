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
	Tick	   time.Duration
}

func GenerateDefaultTimeout() (*TimeoutConfig) {

	var timeout time.Duration = time.Duration(DefaultTimeout) * time.Second
	var tickInterval time.Duration = time.Duration(DefaultTickInterval) * time.Second

	return &TimeoutConfig{
		Timeout: timeout, 
		Tick: tickInterval, 
	}
}

func (tc *TimeoutConfig) NewTimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	// Set new timeout
	tc.Timeout = timeout
	// Set new timeout context and cancel function
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	return ctx, cancel
}

func DoWithTimeout(tc *TimeoutConfig, task func() (interface{}, bool)) (interface{}, error) {
    // Set default values for Timeout and Tick if they're zero
    if tc.Timeout == 0 {
        tc.Timeout = time.Duration(DefaultTimeout) * time.Second
    }
    if tc.Tick == 0 {
        tc.Tick = time.Duration(DefaultTickInterval) * time.Second
    }
    
    ctx, cancel := tc.NewTimeoutContext(tc.Timeout, tc.Tick)
    defer cancel()
    
    ticker := time.NewTicker(tc.Tick)
    defer ticker.Stop()

    for {
        select {
            case <-ctx.Done():
	        return nil, ctx.Err()
	    case <-ticker.C:
		    if result, done := task(); done {
		        return result, nil
		}
	}
    }    
}

func GetPod(tc *TimeoutConfig, clientset *kubernetes.Clientset, namespace string, podName string) (*v1.Pod, bool) {

	pod, err := clientset.CoreV1().Pods(namespace).Get(tc.Ctx, podName, metav1.GetOptions{})

	if err != nil {
		Logger.WithFields(logrus.Fields{
			"namespace": namespace,
			"podName":   podName,
		}).Errorf("Failed to retrieve pod: %v", err)
		return nil, false
	}

	return pod, true
}

