// Package protocol содержит общие протокольные константы и типы.
package protocol

// ErrorCode представляет код ошибки API.
type ErrorCode string

const (
	// ErrorCodeInvalidRequest указывает на невалидный запрос.
	ErrorCodeInvalidRequest ErrorCode = "INVALID_REQUEST"
	// ErrorCodeUnauthorized указывает на отсутствие авторизации.
	ErrorCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	// ErrorCodeForbidden указывает на отсутствие прав доступа.
	ErrorCodeForbidden ErrorCode = "FORBIDDEN"
	// ErrorCodeNotFound указывает на отсутствие ресурса.
	ErrorCodeNotFound ErrorCode = "NOT_FOUND"
	// ErrorCodeConflict указывает на конфликт данных (например, версий).
	ErrorCodeConflict ErrorCode = "CONFLICT"
	// ErrorCodeInternalError указывает на внутреннюю ошибку сервера.
	ErrorCodeInternalError ErrorCode = "INTERNAL_ERROR"
	// ErrorCodeUserExists указывает, что пользователь уже существует.
	ErrorCodeUserExists ErrorCode = "USER_EXISTS"
	// ErrorCodeInvalidCredentials указывает на неверные учетные данные.
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	// ErrorCodeInvalidToken указывает на невалидный токен.
	ErrorCodeInvalidToken ErrorCode = "INVALID_TOKEN"
)

// API версия
const (
	APIVersion = "v1"
	APIPrefix  = "/api/v1"
)

