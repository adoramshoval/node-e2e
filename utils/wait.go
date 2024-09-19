package utils

import (
	"context"
	"time"
)

const (
	DefaultTimeout      int64 = 60
	DefaultTickInterval int64 = 1
)

type TimeoutConfig struct {
	Timeout time.Duration
	Tick    time.Duration
	Ctx     context.Context
	Cancel  context.CancelFunc
}

func GenerateDefaultTimeout() *TimeoutConfig {

	var timeout time.Duration = time.Duration(DefaultTimeout) * time.Second
	var tickInterval time.Duration = time.Duration(DefaultTickInterval) * time.Second

	return &TimeoutConfig{
		Timeout: timeout,
		Tick:    tickInterval,
		Ctx:     nil,
		Cancel:  nil,
	}
}

func (tc *TimeoutConfig) NewTimeoutContext(timeout time.Duration) {
	if timeout <= 0 {
		timeout = time.Duration(DefaultTimeout) * time.Second
	}
	// Set new timeout
	tc.Timeout = timeout
	// Set new timeout context and cancel function
	tc.Ctx, tc.Cancel = context.WithTimeout(context.Background(), timeout)
}

func (tc *TimeoutConfig) DoWithTimeout(task func() (interface{}, bool)) (interface{}, error) {
	// Set default values for Tick if it is zero or lower
	if tc.Tick <= 0 {
		tc.Tick = time.Duration(DefaultTickInterval) * time.Second
	}

	// tc.Timeout should be zero if not set at struct initialization
	tc.NewTimeoutContext(tc.Timeout)

	// Defer cancellation of the context to ensure resources are released
	defer func() {
		tc.Cancel()
		// Reset context and cancel func to nil after use
		tc.Ctx = nil
		tc.Cancel = nil
	}()

	ticker := time.NewTicker(tc.Tick)
	defer ticker.Stop()

	for {
		select {
		case <-tc.Ctx.Done():
			return nil, tc.Ctx.Err()
		case <-ticker.C:
			if result, done := task(); done {
				return result, nil
			}
		}
	}
}
