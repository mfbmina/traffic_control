package circuitbreaker_test

import (
	"net/http"
	"testing"

	"github.com/mfbmina/traffic_control/circuitbreaker"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
)

func Test_Run(t *testing.T) {
	defer apitest.NewMock().
		Get("http://example.com/ok").
		RespondWith().
		Status(http.StatusOK).
		EndStandalone()()

	defer apitest.NewMock().
		Get("http://example.com/fail").
		RespondWith().
		Status(http.StatusBadRequest).
		EndStandalone()()

	t.Run("Circuit breaker should be in CLOSED state if the request is succesful", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker(1)

		req, err := http.NewRequest("GET", "http://example.com/ok", nil)
		assert.NoError(t, err)

		cb.Run(req)
		assert.Equal(t, cb.State, circuitbreaker.CLOSED_STATE)
	})

	t.Run("Circuit breaker should be in CLOSED state if the threshold is not reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker(3)

		req, err := http.NewRequest("GET", "http://example.com/fail", nil)
		assert.NoError(t, err)

		cb.Run(req)
		assert.Equal(t, cb.State, circuitbreaker.CLOSED_STATE)
	})

	t.Run("Circuit breaker should be in OPEN state if the threshold is reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker(3)

		req, err := http.NewRequest("GET", "http://example.com/fail", nil)
		assert.NoError(t, err)

		for range 3 {
			cb.Run(req)
		}

		assert.Equal(t, cb.State, circuitbreaker.OPEN_STATE)
	})
}

func Test_Reset(t *testing.T) {
	cb := circuitbreaker.NewCircuitBreaker(3)
	cb.State = circuitbreaker.OPEN_STATE
	cb.FailureCount = 3

	cb.Reset()
	assert.Equal(t, cb.State, circuitbreaker.CLOSED_STATE)
	assert.Equal(t, cb.FailureCount, 0)
}
