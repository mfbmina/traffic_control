package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mfbmina/traffic_control/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func Test_NewCircuitBreaker(t *testing.T) {
	t.Run("Default config should be used if no option is provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
	})

	t.Run("Should set failure threshold if provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(10)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 10, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
	})

	t.Run("Should set timeout if provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithTimeout(10 * time.Second)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, 10*time.Second, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
	})

	t.Run("Should set success threshold if provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithSuccessThreshold(10)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, 10, cb.SuccessThreshold)
	})
}

func Test_Run(t *testing.T) {
	errorFunc := func() (interface{}, error) { return nil, errors.New("error") }
	successFunc := func() (interface{}, error) { return true, nil }

	t.Run("Circuit breaker should be in CLOSED state if the request is succesful", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		for range 5 {
			cb.Run(successFunc)
		}

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
	})

	t.Run("Circuit breaker should be in CLOSED state if the threshold is not reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		cb.Run(errorFunc)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 1, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state if the threshold is reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		for range 5 {
			cb.Run(errorFunc)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 5, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state while in timeout", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(1)
		cb.Open()
		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)

		for range 5 {
			cb.Run(successFunc)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
	})

	t.Run("Circuit breaker should be in HALF_OPEN state after its timeout", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(1).WithTimeout(1 * time.Second)
		cb.Open()
		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)

		time.Sleep(2 * time.Second)
		cb.Run(successFunc)

		assert.Equal(t, circuitbreaker.HALF_OPEN_STATE, cb.State)
	})

	t.Run("Circuit breaker should be in CLOSED state after succesful requests", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithSuccessThreshold(1)
		cb.HalfOpen()
		assert.Equal(t, circuitbreaker.HALF_OPEN_STATE, cb.State)

		cb.Run(successFunc)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
	})
}
