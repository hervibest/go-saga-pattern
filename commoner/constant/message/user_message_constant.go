package message

const (
	Success = "Successfuly"

	InternalUserAuthNotFound = "user authentication data not found in context"

	ClientInvalidEmailOrPassword = "Make sure you have provide valid email or password"
	ClientUserAlreadyExist       = "Username or email already been used, please use another"
	ClientUnauthenticated        = "Unauthenticated, please try login again"
	ClientPermissionDenied       = "Permission denied for accessing this resource"

	ClientInvalidAccessToken = "Invalid access token, please login again"
)
