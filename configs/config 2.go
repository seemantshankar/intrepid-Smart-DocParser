package configs

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
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
	Dialect      string `mapstructure:"dialect"`
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Name         string `mapstructure:"name"`
	SSLMode      string `mapstructure:"ssl_mode"`
	LogMode      bool   `mapstructure:"log_mode"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleTime  string `mapstructure:"max_idle_time"`
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
		return "postgres"
	}
	return c.Dialect
}

// GetHost returns the database host
func (c *DatabaseConfig) GetHost() string {
	if c.Host == "" {
		return "localhost"
	}
	return c.Host
}

// GetPort returns the database port
func (c *DatabaseConfig) GetPort() string {
	if c.Port == "" {
		return "5432"
	}
	return c.Port
}

// GetUser returns the database user
func (c *DatabaseConfig) GetUser() string {
	if c.User == "" {
		return "postgres"
	}
	return c.User
}

// GetPassword returns the database password
func (c *DatabaseConfig) GetPassword() string {
	return c.Password
}

// GetName returns the database name
func (c *DatabaseConfig) GetName() string {
	if c.Name == "" {
		return "contract_analysis"
	}
	return c.Name
}

// GetSSLMode returns the SSL mode for the database connection
func (c *DatabaseConfig) GetSSLMode() string {
	if c.SSLMode == "" {
		return "disable"
	}
	return c.SSLMode
}

// GetMaxIdleConns returns the maximum number of idle connections
func (c *DatabaseConfig) GetMaxIdleConns() int {
	if c.MaxIdleConns == 0 {
		return 10
	}
	return c.MaxIdleConns
}

// GetMaxOpenConns returns the maximum number of open connections
func (c *DatabaseConfig) GetMaxOpenConns() int {
	if c.MaxOpenConns == 0 {
		return 100
	}
	return c.MaxOpenConns
}

// GetMaxIdleTime returns the maximum idle time for connections
func (c *DatabaseConfig) GetMaxIdleTime() time.Duration {
	if c.MaxIdleTime == "" {
		return 5 * time.Minute
	}
	duration, err := time.ParseDuration(c.MaxIdleTime)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}

// GetLogMode returns whether to enable SQL logging
func (c *DatabaseConfig) GetLogMode() bool {
	return c.LogMode
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.dialect", "sqlite3")
	viper.SetDefault("database.name", "./local.db")
	viper.SetDefault("logging.level", "info")

	// Set config file name and paths
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	
	// Add config paths
	viper.AddConfigPath(".")           // look for config in the working directory
	viper.AddConfigPath("./configs/")  // look for config in the configs directory
	viper.AddConfigPath("$HOME/.config/intrepid-smart-docparser/") // look for config in the user's config directory

	// If a config path is provided, use it
	if configPath != "" {
		viper.SetConfigFile(configPath)
	}

	// Read in environment variables that match
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the config file
	err := viper.ReadInConfig()
	if err != nil {
		// It's okay if the config file doesn't exist, we'll use defaults and environment variables
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal the config into our Config struct
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Override with environment variables if they exist
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		cfg.OCR.APIKey = apiKey
		// Also set it for LLM config if using the same key
		cfg.LLM.OpenRouter.APIKey = apiKey
	}

	return &cfg, nil
}
