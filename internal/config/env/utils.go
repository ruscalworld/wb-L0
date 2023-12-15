package env

import (
	"fmt"
	"os"
)

func requireEnv(key string) string {
	env, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Sprintf("environment variable %s is required, but its value is empty", key))
	}
	return env
}

func envOrDefault(key string, def string) string {
	env, ok := os.LookupEnv(key)
	if ok {
		return env
	}
	return def
}
