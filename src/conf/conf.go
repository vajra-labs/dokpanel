package conf

import (
	"fmt"
	"log"
	"sync"
	"time"

	"dokpanel/src/types"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	NAME        string
	HOST        string `validate:"required"`
	PORT        int    `validate:"required,gt=0,lt=65536"`
	SECRET      string `validate:"required,min=32"`
	GO_ENV      string `validate:"required,oneof=dev prod test"`
	CORS_ORIGIN string `validate:"required"`
	IS_DEV      bool   `validate:"boolean"`
	IS_TEST     bool   `validate:"boolean"`
	IS_PROD     bool   `validate:"boolean"`
	// Rate limiter
	RATE_LIMIT_WINDOWS time.Duration `validate:"required,gt=0"`
	RATE_LIMIT_MAX_REQ int           `validate:"required,gt=0"`
	// Database
	DB_PATH string `validate:"required"`
	// Jwt config
	JWT_ACCESS_EXP  time.Duration `validate:"required,gt=0"`
	JWT_REFRESH_EXP time.Duration `validate:"required,gt=0"`
	// Body parser limit
	BODY_LIMIT int `validate:"required,gt=0"`
	// Docker Config
	DOCKER_HOST        string `validate:"required"`
	DOCKER_API_VERSION string `validate:"required"`
	START_TIME         time.Time
}

var (
	Env     *Config
	once    sync.Once
	VERSION = "dirty" // Overridden with -ldflags
)

func init() {
	once.Do(func() {
		ENV_PATH := getEnv("ENV_PATH", ".env")
		if err := godotenv.Load(ENV_PATH); err != nil {
			log.Printf("Error: %s file not found: %v\n", ENV_PATH, err)
		}
		GO_ENV := getEnv("GO_ENV", types.DEV)
		Env = &Config{
			NAME:               "DokPanel",
			PORT:               getEnvInt("PORT", 8000),
			HOST:               getEnv("HOST", "0.0.0.0"),
			GO_ENV:             GO_ENV,
			CORS_ORIGIN:        getEnv("CORS_ALLOW_ORIGIN"),
			SECRET:             getEnv("SECRET"),
			DB_PATH:            getEnv("DB_PATH"),
			IS_DEV:             GO_ENV == types.DEV,
			IS_TEST:            GO_ENV == types.TEST,
			IS_PROD:            GO_ENV == types.PROD,
			START_TIME:         time.Now(),
			BODY_LIMIT:         int(getEnvByte("BODY_LIMIT")),
			JWT_ACCESS_EXP:     getEnvTime("JWT_ACCESS_EXP", "5m"),
			JWT_REFRESH_EXP:    getEnvTime("JWT_REFRESH_EXP", "24d"),
			RATE_LIMIT_MAX_REQ: getEnvInt("RATE_LIMIT_MAX_REQ", 100),
			RATE_LIMIT_WINDOWS: getEnvTime("RATE_LIMIT_WINDOWS", "15m"),
			DOCKER_HOST:        getEnv("DOCKER_HOST"),
			DOCKER_API_VERSION: getEnv("DOCKER_API_VERSION"),
		}
		if err := validator.New().Struct(Env); err != nil {
			str := fmt.Sprintf("❌ Config validation failed: %v", err)
			panic(str)
		}
	})
}
