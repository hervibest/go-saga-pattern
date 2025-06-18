package errorcode

const (
	// Auth-related
	ErrUnauthorized       string = "UNAUTHORIZED"
	ErrUserSignedOut      string = "USER_SIGNED_OUT"
	ErrInvalidCredentials string = "INVALID_CREDENTIALS"
	ErrForbidden          string = "FORBIDDEN"

	// Validation
	ErrInvalidArgument  string = "INVALID_ARGUMENT"
	ErrMissingField     string = "MISSING_FIELD"
	ErrValidationFailed string = "VALIDATION_FAILED"

	// Not found
	ErrUserNotFound     string = "USER_NOT_FOUND"
	ErrResourceNotFound string = "RESOURCE_NOT_FOUND"

	// Conflict
	ErrAlreadyExists string = "ALREADY_EXISTS"

	// Internal
	ErrInternal        string = "INTERNAL"
	ErrCacheFailure    string = "CACHE_FAILURE"
	ErrDatabaseFailure string = "DATABASE_FAILURE"

	// External
	ErrExternal string = "EXTERNAL"

	// Rate limit
	ErrTooManyRequests string = "TOO_MANY_REQUESTS"
)
