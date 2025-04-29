package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mfbmina/traffic_control/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func Test_New(t *testing.T) {
	t.Run("Default config should be used if no option is provided", func(t *testing.T) {
		cb := circuitbreaker.New()

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set failure threshold if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithFailureThreshold(10)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 10, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set timeout if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithTimeout(10 * time.Second)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, 10*time.Second, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set success threshold if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithSuccessThreshold(10)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, 10, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set close check if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithCloseCheck(func(cb circuitbreaker.CircuitBreaker) bool { return true })

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.True(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set half open check if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithHalfOpenCheck(func(cb circuitbreaker.CircuitBreaker) bool { return true })

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.True(t, cb.HalfOpenCheck(*cb))
		assert.False(t, cb.OpenCheck(*cb))
	})

	t.Run("Should set open check if provided", func(t *testing.T) {
		cb := circuitbreaker.New().WithOpenCheck(func(cb circuitbreaker.CircuitBreaker) bool { return true })

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
		assert.Equal(t, circuitbreaker.DEFAULT_THRESHOLD, cb.SuccessThreshold)
		assert.False(t, cb.CloseCheck(*cb))
		assert.False(t, cb.HalfOpenCheck(*cb))
		assert.True(t, cb.OpenCheck(*cb))
	})
}

func Test_Run(t *testing.T) {
	errorFunc := func() (interface{}, error) { return nil, errors.New("error") }
	successFunc := func() (interface{}, error) { return true, nil }

	t.Run("Circuit breaker should be in CLOSED state if the request is successful", func(t *testing.T) {
		cb := circuitbreaker.New()

		for range 5 {
			r, err := cb.Run(successFunc)
			assert.True(t, r.(bool))
			assert.NoError(t, err)
		}

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
	})

	t.Run("Circuit breaker should be in CLOSED state if the threshold is not reached", func(t *testing.T) {
		cb := circuitbreaker.New()

		r, err := cb.Run(errorFunc)
		assert.Nil(t, r)
		assert.NotNil(t, err)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 1, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state if the threshold is reached", func(t *testing.T) {
		cb := circuitbreaker.New()

		for range 5 {
			r, err := cb.Run(errorFunc)
			assert.Nil(t, r)
			assert.NotNil(t, err)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 5, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state while in timeout", func(t *testing.T) {
		cb := circuitbreaker.New().WithFailureThreshold(1)
		cb.Open()
		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)

		for range 5 {
			r, err := cb.Run(successFunc)
			assert.Nil(t, r)
			assert.ErrorIs(t, err, circuitbreaker.ErrOpenCircuit)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
	})

	t.Run("Circuit breaker should be in HALF_OPEN state after its timeout", func(t *testing.T) {
		cb := circuitbreaker.New().WithFailureThreshold(1).WithTimeout(1 * time.Second)
		cb.Open()
		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)

		time.Sleep(2 * time.Second)
		r, err := cb.Run(successFunc)
		assert.NotNil(t, r)
		assert.NoError(t, err)
		assert.Equal(t, circuitbreaker.HALF_OPEN_STATE, cb.State)
	})

	t.Run("Circuit breaker should be in CLOSED state after successful requests", func(t *testing.T) {
		cb := circuitbreaker.New().WithSuccessThreshold(1)
		cb.HalfOpen()
		assert.Equal(t, circuitbreaker.HALF_OPEN_STATE, cb.State)

		r, err := cb.Run(successFunc)
		assert.NotNil(t, r)
		assert.NoError(t, err)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
	})
}
