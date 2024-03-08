package env

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

const Env = "ENV"

func GetEnvWithFallback(env string, fallback string) string {
	value := os.Getenv(env)
	if value == "" {
		value = fallback
	}
	return value
}

func GetEnvOrFail(env string) string {
	value := os.Getenv(env)
	if value == "" {
		zap.L().Fatal(fmt.Sprintf("Env var %s not set", env), zap.String("env", env))
	}
	return value
}

func IsProd() bool {
	return os.Getenv(Env) == "prod"
}

func IsLocal() bool {
	env := os.Getenv(Env)
	return env == "" || env == "local"
}
