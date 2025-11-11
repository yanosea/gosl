package errorhandler

import (
	"errors"
	"fmt"
)

type ErrorCategory int

const (
	ErrorCategoryAuth ErrorCategory = iota
	ErrorCategoryNetwork
	ErrorCategoryRateLimit
	ErrorCategoryValidation
	ErrorCategoryUnknown
)

func (ec ErrorCategory) String() string {
	switch ec {
	case ErrorCategoryAuth:
		return "Auth"
	case ErrorCategoryNetwork:
		return "Network"
	case ErrorCategoryRateLimit:
		return "RateLimit"
	case ErrorCategoryValidation:
		return "Validation"
	case ErrorCategoryUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrNetworkError      = errors.New("network error")
	ErrChannelNotFound   = errors.New("channel not found")
)

type AppError struct {
	Category    ErrorCategory
	Message     string
	Recoverable bool
	Err         error
}

func (e AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e AppError) Unwrap() error {
	return e.Err
}

func HandleError(err error) AppError {
	if err == nil {
		return AppError{
			Category:    ErrorCategoryUnknown,
			Message:     "unknown error occurred",
			Recoverable: false,
			Err:         errors.New("nil error"),
		}
	}

	switch {
	case errors.Is(err, ErrInvalidToken):
		return AppError{
			Category:    ErrorCategoryAuth,
			Message:     "invalid slack token, please check configuration file",
			Recoverable: false,
			Err:         err,
		}
	case errors.Is(err, ErrRateLimitExceeded):
		return AppError{
			Category:    ErrorCategoryRateLimit,
			Message:     "api rate limit exceeded, please wait",
			Recoverable: true,
			Err:         err,
		}
	case errors.Is(err, ErrNetworkError):
		return AppError{
			Category:    ErrorCategoryNetwork,
			Message:     "network error occurred, please check connection",
			Recoverable: true,
			Err:         err,
		}
	case errors.Is(err, ErrChannelNotFound):
		return AppError{
			Category:    ErrorCategoryValidation,
			Message:     "specified channel not found",
			Recoverable: false,
			Err:         err,
		}
	default:
		return AppError{
			Category:    ErrorCategoryUnknown,
			Message:     "unknown error occurred, check logs for details",
			Recoverable: false,
			Err:         err,
		}
	}
}
