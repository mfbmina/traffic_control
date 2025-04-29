# Circuit breaker

Implements a simple circuit breaker in Go. Inspired by [gobreaker](https://github.com/sony/gobreaker)

## Usage

Using is simple, just initialize a new circuit breaker and then wrap your function inside of it. Please notice that the function must follow the contract `func() (interface{}, error)`.

```go
func main() {
cb := circuitbreaker.New()

cb.Run(myfunc)
}

func myFunc() (interface{}, error) {
  return true, nil
} 
```

For each error, a counter is increased until it reaches the threshold and opens the circuit.

## Configuration

You can configure some options:

```go
func main() {
cb := circuitbreaker.New().
        WithFailureThreshold(100) // errors threshold before transitioning from CLOSED to OPEN
        WithSuccessThreshold(100) // success threshold before transitioning from HALF_OPEN to CLOSED
        WithTimeout(10 * time.Second) // how long it waits before transitioning from OPEN to HALF_OPEN

cb.Run(myfunc)
}

func myFunc() (interface{}, error) {
  return true, nil
} 
```

## Transitioning
If you wish, you can change the state by yourself.

```go
cb := circuitbreaker.New()
cb.Open()
cb.HalfOpen()
cb.Close()
```

