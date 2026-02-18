package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Port      string `mapstructure:"PORT"`
	DBConn    string `mapstructure:"DB_CONN"`
	AppEnv    string `mapstructure:"APP_ENV"`
	AppURL    string `mapstructure:"APP_URL"`
	JWTSecret string `mapstructure:"JWT_SECRET"`
}

// LoadConfig reads configuration from environment variables and optional .env file
func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	cfg := &Config{
		Port:      viper.GetString("PORT"),
		DBConn:    viper.GetString("DB_CONN"),
		AppEnv:    viper.GetString("APP_ENV"),
		AppURL:    viper.GetString("APP_URL"),
		JWTSecret: viper.GetString("JWT_SECRET"),
	}

	// Defaults
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = "change-me-in-production"
	}

	return cfg, nil
}

// IsProduction returns true if APP_ENV is "production"
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// SwaggerHost returns the host for Swagger documentation
func (c *Config) SwaggerHost() string {
	if c.AppURL != "" {
		return c.AppURL
	}
	return "localhost:" + c.Port
}

// SwaggerSchemes returns the schemes for Swagger documentation
func (c *Config) SwaggerSchemes() []string {
	if c.IsProduction() {
		return []string{"https"}
	}
	return []string{"http"}
}
