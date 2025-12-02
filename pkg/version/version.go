// Package version предоставляет информацию о версии и дате сборки приложения.
// Значения устанавливаются через build flags при компиляции.
package version

var (
	// Version содержит версию приложения (например, "1.0.0").
	// Устанавливается через -ldflags "-X github.com/victor/gophkeeper2/pkg/version.Version=1.0.0"
	Version = "dev"

	// BuildDate содержит дату сборки в формате RFC3339.
	// Устанавливается через -ldflags "-X github.com/victor/gophkeeper2/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildDate = "unknown"
)

// GetVersion возвращает версию приложения.
func GetVersion() string {
	return Version
}

// GetBuildDate возвращает дату сборки приложения.
func GetBuildDate() string {
	return BuildDate
}

// GetInfo возвращает строку с информацией о версии и дате сборки.
func GetInfo() string {
	return "Version: " + Version + "\nBuild Date: " + BuildDate
}

