package storage

import "errors"

var (
	// ErrUserNotFound возвращается, когда пользователь не найден.
	ErrUserNotFound = errors.New("user not found")
	// ErrUserExists возвращается, когда пользователь уже существует.
	ErrUserExists = errors.New("user already exists")
	// ErrSessionNotFound возвращается, когда сессия не найдена.
	ErrSessionNotFound = errors.New("session not found")
	// ErrRecordNotFound возвращается, когда запись данных не найдена.
	ErrRecordNotFound = errors.New("record not found")
	// ErrVersionConflict возвращается при конфликте версий при обновлении.
	ErrVersionConflict = errors.New("version conflict")
)

