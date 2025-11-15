package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration settings.
type Config struct {

	// Infrastructure settings (Crucial)
	DBUrl   string `mapstructure:"DB_URL"`
	APIPort string `mapstructure:"API_PORT"`
	WSPort  string `mapstructure:"WS_PORT"`

	// Server settings
	ReadTimeout       time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	ReadHeaderTimeout time.Duration `mapstructure:"SERVER_READ_HEADER_TIMEOUT"`
	WriteTimeout      time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`

	// Security settings
	JWTSecret    string        `mapstructure:"JWT_SECRET"`
	TokenTimeout time.Duration `mapstructure:"TOKEN_TIMEOUT"`
}

// LoadConfig initializes Viper and loads configuration.
func LoadConfig() Config {
	// Set default values
	viper.SetDefault("WS_PORT", "8080")
	viper.SetDefault("API_PORT", "8081")
	viper.SetDefault("SERVER_READ_TIMEOUT", 10*time.Second)
	viper.SetDefault("SERVER_READ_HEADER_TIMEOUT", 5*time.Second)
	viper.SetDefault("SERVER_WRITE_TIMEOUT", 10*time.Second)
	viper.SetDefault("TOKEN_TIMEOUT", 24*time.Hour)
	err := viper.BindEnv("DB_URL")
	if err != nil {
		return Config{}
	}

	err = viper.BindEnv("JWT_SECRET")
	if err != nil {
		return Config{}
	}

	// Read environment variables (e.g., JWT_SECRET=...)
	viper.AutomaticEnv()

	// Ensure the critical secret is present
	if viper.GetString("JWT_SECRET") == "" {
		log.Fatal("FATAL ERROR: JWT_SECRET environment variable is required and must not be empty.")
	}

	if viper.GetString("DB_URL") == "" {
		log.Fatal("FATAL ERROR: DB_URL environment variable is required and must not be empty.")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("FATAL ERROR: Failed to unmarshal config: %v", err)
	}

	return cfg
}
