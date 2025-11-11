package errorhandler

import (
	"errors"
	"testing"
)

// TestAppError_Error tests that AppError implements the error interface correctly
func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		appErr  AppError
		want    string
	}{
		{
			name: "Auth error with Japanese message",
			appErr: AppError{
				Category:    ErrorCategoryAuth,
				Message:     "認証に失敗しました",
				Recoverable: false,
				Err:         errors.New("invalid token"),
			},
			want: "認証に失敗しました: invalid token",
		},
		{
			name: "Network error without underlying error",
			appErr: AppError{
				Category:    ErrorCategoryNetwork,
				Message:     "ネットワークエラーが発生しました",
				Recoverable: true,
				Err:         nil,
			},
			want: "ネットワークエラーが発生しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			if got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAppError_Unwrap tests error unwrapping
func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	appErr := AppError{
		Category:    ErrorCategoryNetwork,
		Message:     "ネットワークエラー",
		Recoverable: true,
		Err:         underlyingErr,
	}

	unwrapped := appErr.Unwrap()
	if !errors.Is(unwrapped, underlyingErr) {
		t.Errorf("AppError.Unwrap() did not return the underlying error")
	}
}

// TestHandleError tests error categorization
func TestHandleError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		wantCategory     ErrorCategory
		wantRecoverable  bool
		wantMessagePart  string
	}{
		{
			name:             "Invalid token error",
			err:              ErrInvalidToken,
			wantCategory:     ErrorCategoryAuth,
			wantRecoverable:  false,
			wantMessagePart:  "トークン",
		},
		{
			name:             "Rate limit exceeded error",
			err:              ErrRateLimitExceeded,
			wantCategory:     ErrorCategoryRateLimit,
			wantRecoverable:  true,
			wantMessagePart:  "制限",
		},
		{
			name:             "Network error",
			err:              ErrNetworkError,
			wantCategory:     ErrorCategoryNetwork,
			wantRecoverable:  true,
			wantMessagePart:  "ネットワーク",
		},
		{
			name:             "Channel not found error",
			err:              ErrChannelNotFound,
			wantCategory:     ErrorCategoryValidation,
			wantRecoverable:  false,
			wantMessagePart:  "チャンネル",
		},
		{
			name:             "Unknown error",
			err:              errors.New("some random error"),
			wantCategory:     ErrorCategoryUnknown,
			wantRecoverable:  false,
			wantMessagePart:  "不明なエラー",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := HandleError(tt.err)

			if appErr.Category != tt.wantCategory {
				t.Errorf("HandleError() Category = %v, want %v", appErr.Category, tt.wantCategory)
			}

			if appErr.Recoverable != tt.wantRecoverable {
				t.Errorf("HandleError() Recoverable = %v, want %v", appErr.Recoverable, tt.wantRecoverable)
			}

			if appErr.Message == "" {
				t.Error("HandleError() Message is empty")
			}

			// Check if the message contains expected Japanese text
			if tt.wantMessagePart != "" {
				// Simple substring check for Japanese message
				found := false
				for _, r := range tt.wantMessagePart {
					for _, mr := range appErr.Message {
						if r == mr {
							found = true
							break
						}
					}
				}
				if !found && len(appErr.Message) < len(tt.wantMessagePart) {
					t.Errorf("HandleError() Message = %v, expected to contain %v", appErr.Message, tt.wantMessagePart)
				}
			}

			if appErr.Err == nil {
				t.Error("HandleError() underlying Err should not be nil")
			}
		})
	}
}

// TestHandleError_WrappedErrors tests handling of wrapped errors
func TestHandleError_WrappedErrors(t *testing.T) {
	wrappedErr := errors.Join(ErrNetworkError, errors.New("connection timeout"))
	appErr := HandleError(wrappedErr)

	if appErr.Category != ErrorCategoryNetwork {
		t.Errorf("HandleError() with wrapped error Category = %v, want %v", appErr.Category, ErrorCategoryNetwork)
	}

	if !errors.Is(appErr.Err, wrappedErr) {
		t.Error("HandleError() should preserve wrapped error chain")
	}
}

// TestErrorCategory_String tests error category string representation
func TestErrorCategory_String(t *testing.T) {
	tests := []struct {
		name     string
		category ErrorCategory
		want     string
	}{
		{"Auth category", ErrorCategoryAuth, "Auth"},
		{"Network category", ErrorCategoryNetwork, "Network"},
		{"RateLimit category", ErrorCategoryRateLimit, "RateLimit"},
		{"Validation category", ErrorCategoryValidation, "Validation"},
		{"Unknown category", ErrorCategoryUnknown, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.category.String()
			if got != tt.want {
				t.Errorf("ErrorCategory.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
