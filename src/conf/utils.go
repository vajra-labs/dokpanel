package conf

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

func getEnv(key string, fallback ...string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	panic(fmt.Sprintf("Env var %s not found and no fallback provided", key))
}

func getEnvInt[T int | int64](key string, fallback ...T) T {
	str := os.Getenv(key)
	if str == "" {
		if len(fallback) > 0 {
			return fallback[0]
		}
		panic(fmt.Sprintf("Env var %s not found and no fallback provided", key))
	}
	parsed, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		panic(fmt.Sprintf("Failed to parse %s: %v", key, err))
	}
	return T(parsed)
}

func getEnvByte(key string) uint64 {
	str := getEnv(key)
	bytes, err := humanize.ParseBytes(str)
	if err != nil {
		panic(err)
	}
	return bytes
}

func getEnvTime(key string, fallback ...string) time.Duration {
	str := getEnv(key, fallback...)
	d, err := time.ParseDuration(str)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse %s: %v", key, err))
	}
	return d
}

func getEnvBool(key string, fallback bool) bool {
	val := getEnv(key)
	if val == "" {
		return fallback
	}
	return val == "true" || val == "1"
}
