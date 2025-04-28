package circuitbreaker_test

import (
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/mfbmina/traffic_control/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

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
		assert.Equal(t, 0, cb.FailureCount)
	})

	t.Run("Circuit breaker should be in CLOSED state if the threshold is not reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)

		cb.Run(req)
		assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
		assert.Equal(t, 1, cb.FailureCount)
	})

	t.Run("Circuit breaker should be in OPEN state if the threshold is reached", func(t *testing.T) {
		cb := circuitbreaker.NewCircuitBreaker()

		req, err := http.NewRequest("GET", "http://example.com/nok", nil)
		assert.NoError(t, err)

		for range 5 {
			cb.Run(req)
		}

		assert.Equal(t, circuitbreaker.OPEN_STATE, cb.State)
		assert.Equal(t, 5, cb.FailureCount)
	})
}

func Test_Reset(t *testing.T) {
	cb := circuitbreaker.NewCircuitBreaker()
	cb.State = circuitbreaker.OPEN_STATE
	cb.FailureCount = 3

	cb.Reset()
	assert.Equal(t, circuitbreaker.CLOSED_STATE, cb.State)
	assert.Equal(t, 0, cb.FailureCount)
}
