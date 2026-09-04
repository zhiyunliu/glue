package circuitbreaker

//type CircuitBreaker = circuitbreaker.CircuitBreaker
type CircuitBreaker interface {
	Allow() error
	MarkSuccess()
	MarkFailed()
}
