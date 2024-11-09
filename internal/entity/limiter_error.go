package entity

type LimiterError struct {
	Message string
	Err     string
}

func (le *LimiterError) Error() string {
	return le.Message
}

func NewExpiredLimiterError() *LimiterError {
	return &LimiterError{
		Message: "The rate limit has expired and is no longer valid.",
		Err:     "expired_limiter",
	}
}

func NewIncrementBlockedError() *LimiterError {
	return &LimiterError{
		Message: "You have exceeded the allowed number of requests within the specified time window and have been temporarily blocked.",
		Err:     "is_blocked",
	}
}

func NewEntityNotFound() *LimiterError {
	return &LimiterError{
		Message: "The requested entity could not be found in the system.",
		Err:     "entity_not_found",
	}
}
