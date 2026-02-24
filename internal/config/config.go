package config

import (
	"os"
	"strconv"
)

type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string

	JWTKey        string
	JWTExpireHour float64

	SMTPHost string
	SMTPPort string
	SMTPUser string
	SMTPPass string
	SMTPFrom string
}

func Load() *Config {
	return &Config{
		DBHost:        getEnv("DB_HOST", ""),
		DBPort:        getEnv("DB_PORT", ""),
		DBUser:        getEnv("DB_USER", ""),
		DBPass:        getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", ""),
		DBSSLMode:     getEnv("DB_SSLMODE", ""),
		JWTKey:        getEnv("JWT_SECRET", ""),
		JWTExpireHour: getEnvFloat("JWT_EXPIRE_HOUR", 24),
		SMTPHost:      getEnv("SMTP_HOST", ""),
		SMTPPort:      getEnv("SMTP_PORT", ""),
		SMTPUser:      getEnv("SMTP_USER", ""),
		SMTPPass:      getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:      getEnv("SMTP_FROM", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(k string, defaultV int) int {
	v := os.Getenv(k)
	if v == "" {
		return defaultV
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultV
	}
	return n
}

func getEnvFloat(k string, defaultV float64) float64 {
	v := os.Getenv(k)
	if v == "" {
		return defaultV
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultV
	}
	return n
}
