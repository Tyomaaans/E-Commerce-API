package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"E-COMMERCE-API/internal/infrastructure/redpanda"
)

type AppConfig struct {
	APPport              string
	DSN                  string
	JWTSecretKey         string
	JWTExpiry            time.Duration
    DefaultRefreshExpiry time.Duration
	ShortRefreshExpiry   time.Duration
	REDISaddr            string
	REDISpassword        string
	SMTPhost             string
    SMTPport             string
    SMTPfrom             string
    SMTPusername         string
    SMTPpassword         string
    APPurl               string
	REDPANDAbrokers      []string
}

func NewConfig() AppConfig {
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found!")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
        log.Fatal("APP_PORT environment variable is required!")
    }

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environtment variable is required!")
	}

	accessExpiryStr := os.Getenv("JWT_EXPIRY")
	accessExpiry, err := time.ParseDuration(accessExpiryStr)
	if err != nil {
		log.Fatal("invalid JWT_EXPIRY format!")
	}

    defaultExpiryStr := os.Getenv("DEFAULT_REFRESH_EXPIRY")
	defaultExpiry, err := time.ParseDuration(defaultExpiryStr)
	if err != nil {
		log.Fatal("invalid DEFAULT_REFRESH_EXPIRY format!")
	}

	shortExpiryStr := os.Getenv("SHORT_REFRESH_EXPIRY")
	shortExpiry, err := time.ParseDuration(shortExpiryStr)
	if err != nil {
		log.Fatal("invalid DEFAULT_REFRESH_EXPIRY format!")
	} 
	
    secretKey := os.Getenv("JWT_SECRET_KEY")
    if secretKey == "" {
        log.Fatal("JWT_SECRET_KEY environment variable is required!")
    }

	redisAddr := os.Getenv("REDIS_ADDR")
    if redisAddr == "" {
        log.Fatal("REDIS_ADDR environment variable is required!")
    }

	redisPassword := os.Getenv("REDIS_PASSWORD")
	if redisPassword == "" {
        log.Fatal("REDIS_PASSWORD environment variable is required!")
    }

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		log.Fatal("SMTP_HOST environment variable is required!")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		log.Fatal("SMTP_PORT environment variable is required!")
	}

	smtpFrom := os.Getenv("SMTP_FROM")
	if smtpFrom == "" {
		log.Fatal("SMTP_FROM environment variable is required!")
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		log.Fatal("APP_URL environment variable is required!")
	}

	redpandaBrokers := os.Getenv("REDPANDA_BROKERS")
	if redpandaBrokers == "" {
		log.Fatal("REDPANDA_BROKERS environment variable is required!")
	}

	return AppConfig{
		APPport:              appPort,
		DSN:                  dsn,
		JWTSecretKey:         secretKey,
        JWTExpiry:            accessExpiry,
        DefaultRefreshExpiry: defaultExpiry,
		ShortRefreshExpiry:   shortExpiry,
		REDISaddr:            redisAddr,
		REDISpassword:        redisPassword,
		SMTPhost:             smtpHost,
		SMTPport:             smtpPort,
		SMTPfrom:             smtpFrom,
		SMTPusername:         os.Getenv("SMTP_USERNAME"),
		SMTPpassword:         os.Getenv("SMTP_PASSWORD"),
		APPurl:               appURL,
		REDPANDAbrokers:      redpanda.ParseBrokers(redpandaBrokers),
	}
}