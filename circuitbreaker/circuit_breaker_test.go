package circuitbreaker_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/mfbmina/traffic_control/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

func Test_NewCircuitBreaker(t *testing.T) {
	t.Run("Default config should be used if no option is provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
		assert.Equal(t, circuitbreaker.DEFAULT_FAILURE_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
	})

	t.Run("Should set failure threshold if provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(10)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
		assert.Equal(t, 10, cb.FailureThreshold)
		assert.Equal(t, circuitbreaker.DEFAULT_TIMEOUT, cb.Timeout)
	})

	t.Run("Should set timeout if provided", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithTimeout(10 * time.Second)

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
		assert.Equal(t, circuitbreaker.DEFAULT_FAILURE_THRESHOLD, cb.FailureThreshold)
		assert.Equal(t, 10*time.Second, cb.Timeout)
	})
}

func Test_Run(t *testing.T) {
	t.Run("Circuit breaker should be in CLOSED state if the request is succesful", func(t *testing.T) {
		defer gock.Off()

		cb := circuitbreaker.NewCircuitBreaker()

		req, err := http.NewRequest("GET", "http://example.com/ok", nil)
		assert.NoError(t, err)

		for range 5 {
			gock.New("http://example.com").
				Get("/ok").
				Reply(http.StatusOK)

			cb.Run(req)
		}

		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
	})

	t.Run("Circuit breaker should be in CLOSED state if the threshold is not reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)

		cb.Run(req)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 1, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state if the threshold is reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)

		for range 5 {
			cb.Run(req)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 5, cb.Failures)
	})

	t.Run("Circuit breaker should be in OPEN state while in timeout", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(1)

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)

		for range 5 {
			cb.Run(req)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 1, cb.Failures)
	})

	t.Run("Circuit breaker should be in HALF_OPEN state after its timeout", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker().WithFailureThreshold(1).WithTimeout(1 * time.Second)

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)
		cb.Run(req)
		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 1, cb.Failures)

		time.Sleep(2 * time.Second)
		gock.New("http://example.com").Get("/ok").Reply(http.StatusOK)
		req, err = http.NewRequest("GET", "http://example.com/ok", nil)
		assert.NoError(t, err)
		cb.Run(req)
		assert.Equal(t, circuitbreaker.HALF_OPEN_STATE, cb.State)
		assert.Equal(t, 0, cb.Failures)
	})
}
