package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	RedisAddr      string
	RedisPassword  string
	GoogleClientID string
	JWTSecret      string
	SeatLockTTL    time.Duration
	AllowedOrigins string
	AdminEmails    []string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from system environment")
	}
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDBName:    getEnv("MONGO_DB_NAME", "cinema_booking"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "http://localhost:5173"),
	}

	seatLockSeconds, err := strconv.Atoi(getEnv("SEAT_LOCK_TTL_SECONDS", "300"))
	if err != nil || seatLockSeconds <= 0 {
		log.Println("WARNING: invalid SEAT_LOCK_TTL_SECONDS, falling back to 300")
		seatLockSeconds = 300
	}
	cfg.SeatLockTTL = time.Duration(seatLockSeconds) * time.Second

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required and must not be empty — refusing to start with an insecure default")
	}

	adminEmailsRaw := getEnv("ADMIN_EMAILS", "")
	if adminEmailsRaw != "" {
		for _, e := range strings.Split(adminEmailsRaw, ",") {
			e = strings.TrimSpace(strings.ToLower(e))
			if e != "" {
				cfg.AdminEmails = append(cfg.AdminEmails, e)
			}
		}
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
