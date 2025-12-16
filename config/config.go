package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort  string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	Environment string
	Scraper     ScraperConfig
}

type ScraperConfig struct {
	MaxScrolls              int
	MaxDuration             time.Duration
	InitialDelay            time.Duration
	MinDelay                time.Duration
	MaxDelay                time.Duration
	MaxConsecutiveNoNew     int
	MaxConsecutiveUnchanged int
	ExtractionInterval      int
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "3001"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBName:      getEnv("DB_NAME", "car_listing"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Scraper:     loadScraperConfig(),
	}
}

func loadScraperConfig() ScraperConfig {
	return ScraperConfig{
		MaxScrolls:              getEnvInt("SCRAPER_MAX_SCROLLS", 2000),
		MaxDuration:             getEnvDuration("SCRAPER_MAX_DURATION", 60*time.Minute),
		InitialDelay:            getEnvDuration("SCRAPER_INITIAL_DELAY", 2*time.Second),
		MinDelay:                getEnvDuration("SCRAPER_MIN_DELAY", 1500*time.Millisecond),
		MaxDelay:                getEnvDuration("SCRAPER_MAX_DELAY", 5*time.Second),
		MaxConsecutiveNoNew:     getEnvInt("SCRAPER_MAX_CONSECUTIVE_NO_NEW", 10),
		MaxConsecutiveUnchanged: getEnvInt("SCRAPER_MAX_CONSECUTIVE_UNCHANGED", 10),
		ExtractionInterval:      getEnvInt("SCRAPER_EXTRACTION_INTERVAL", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	if defaultValue == "" {
		log.Printf("Warning: %s environminent variable not set", key)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: %s is not a valid integer, using default: %d", key, defaultValue)
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: %s is not a valid duration, using default: %v", key, defaultValue)
	}
	return defaultValue
}
