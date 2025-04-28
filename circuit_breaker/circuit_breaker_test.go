package circuitbreaker_test

import (
	"net/http"
	"testing"
)

func TestRun(t *testing.T) {
	failureFunc := func(r *http.Response) bool {
		// Simulate a failure if the status code is not 200
		return r.StatusCode != http.StatusOK
	}

	cases := []struct {
		cb            *circuitbreaker.CircuitBreaker
		expectedState string
		message       string
	}{
		{
			cb:            circuitbreaker.NewCircuitBreaker(3, failureFunc),
			expectedState: circuitbreaker.CLOSED_STATE,
			message:       "Circuit breaker should be in CLOSED state after one failure",
		},
		{
			cb:            circuitbreaker.NewCircuitBreaker(3, failureFunc),
			expectedState: circuitbreaker.CLOSED_STATE,
			message:       "Circuit breaker should be in OPEN state after three failure",
		},
	}
	// assert equality

	for _, c := range cases {
		got := ReverseRunes(c.in)
		if got != c.want {
			t.Errorf("ReverseRunes(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}
