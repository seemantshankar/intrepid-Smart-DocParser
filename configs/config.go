package configs

import (
	"github.com/spf13/viper"
	"os"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment string         `mapstructure:"environment"`
	ServiceName string         `mapstructure:"service_name"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Jaeger      JaegerConfig   `mapstructure:"jaeger"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	LLM         LLMConfig      `mapstructure:"llm"`
	OCR         OCRConfig      `mapstructure:"ocr"`
	Redis       RedisConfig    `mapstructure:"redis"`
}

// LLMConfig holds configuration for all LLM providers
type LLMConfig struct {
	OpenRouter LLMProviderConfig `mapstructure:"openrouter"`
}

// OCRConfig holds configuration for the OCR provider
type OCRConfig struct {
	APIKey         string   `mapstructure:"api_key"`
	FallbackModels []string `mapstructure:"fallback_models"`
}

// RedisConfig holds configuration for Redis
type RedisConfig struct {
	Address  string `mapstructure:"address"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LLMProviderConfig holds configuration for a single LLM provider
type LLMProviderConfig struct {
	BaseURL       string        `mapstructure:"base_url"`
	APIKey        string        `mapstructure:"api_key"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryCount    int           `mapstructure:"retry_count"`
	RetryWaitTime time.Duration `mapstructure:"retry_wait_time"`
	RetryMaxInterval time.Duration `mapstructure:"retry_max_interval"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string        `mapstructure:"port" default:"9091"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Dialect string `mapstructure:"dialect"`
	Name   string `mapstructure:"name"`
	LogMode bool   `mapstructure:"log_mode"`
}

// JaegerConfig holds Jaeger tracing configuration
type JaegerConfig struct {
	URL          string  `mapstructure:"url"`
	SamplingRate float64 `mapstructure:"sampling_rate"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level string `mapstructure:"level"`
}

// GetDialect returns the database dialect
func (c DatabaseConfig) GetDialect() string {
	if c.Dialect == "" {
		return "sqlite3" // Default to SQLite if not specified
	}
	return c.Dialect
}

// GetName returns the database name or file path
func (c DatabaseConfig) GetName() string {
	if c.Name == "" {
		return "./local.db" // Default SQLite database file
	}
	return c.Name
}

// GetLogMode returns whether to enable SQL logging
func (c DatabaseConfig) GetLogMode() bool {
	return c.LogMode
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	
	// Override config with environment variables if they exist
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		cfg.OCR.APIKey = apiKey
	}
	
	return &cfg, nil
}
