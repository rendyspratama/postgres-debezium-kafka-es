package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App            AppConfig            `yaml:"app"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	ES             ElasticsearchConfig  `yaml:"es"`
	Sync           SyncConfig           `yaml:"sync"`
	Monitoring     MonitoringConfig     `yaml:"monitoring"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
}

type AppConfig struct {
	Environment string `yaml:"environment"`
	LogLevel    string `yaml:"log_level"`
	ServiceName string `yaml:"service_name"`
	Version     string `yaml:"version"`
}

type KafkaConfig struct {
	Brokers         []string `yaml:"brokers"`
	GroupID         string   `yaml:"group_id"`
	TopicPrefix     string   `yaml:"topic_prefix"`
	AutoOffsetReset string   `yaml:"auto_offset_reset"`
	SecurityEnabled bool     `yaml:"security_enabled"`
	SASL            struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"sasl"`
	// Security configs to be added later
}

type ElasticsearchConfig struct {
	Hosts       []string      `yaml:"hosts"`
	IndexPrefix string        `yaml:"index_prefix"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	MaxRetries  int           `yaml:"max_retries"`
	Timeout     time.Duration `yaml:"timeout"`
	// Add more ES-specific configs
	MaxConns       int           `yaml:"max_conns"`
	MaxIdleConns   int           `yaml:"max_idle_conns"`
	ConnectTimeout time.Duration `yaml:"connect_timeout"`
	RequestTimeout time.Duration `yaml:"request_timeout"`
	RetryBackoff   time.Duration `yaml:"retry_backoff"`
	EnableRetry    bool          `yaml:"enable_retry"`
	EnableMetrics  bool          `yaml:"enable_metrics"`
	SnifferEnabled bool          `yaml:"sniffer_enabled"`
	GzipEnabled    bool          `yaml:"gzip_enabled"`

	// Index naming strategy
	IndexTemplate  string `yaml:"index_template"`
	IndexLifecycle string `yaml:"index_lifecycle"`
	ShardCount     int    `yaml:"shard_count"`
	ReplicaCount   int    `yaml:"replica_count"`
}

type SyncConfig struct {
	Mode         string             `yaml:"mode"`
	KafkaConnect KafkaConnectConfig `yaml:"kafka_connect"`
	Custom       CustomConfig       `yaml:"custom"`
}

type KafkaConnectConfig struct {
	Enabled       bool                `yaml:"enabled"`
	SinkConnector SinkConnectorConfig `yaml:"sink_connector"`
}

type SinkConnectorConfig struct {
	URL         string `yaml:"url"`
	Name        string `yaml:"name"`
	TopicPrefix string `yaml:"topic_prefix"`
}

type CustomConfig struct {
	Enabled       bool          `yaml:"enabled"`
	BatchSize     int           `yaml:"batch_size"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryDelay    time.Duration `yaml:"retry_delay"`
	MaxRetryDelay time.Duration `yaml:"max_retry_delay"`
	BackoffFactor float64       `yaml:"backoff_factor"`
	FailureQueue  string        `yaml:"failure_queue"`
	ConflictMode  string        `yaml:"conflict_mode"`
}

type MonitoringConfig struct {
	Enabled        bool `yaml:"enabled"`
	MetricsPort    int  `yaml:"metrics_port"`
	TracingEnabled bool `yaml:"tracing_enabled"`
	// OpenTelemetry configuration
	OtelCollector string `yaml:"otel_collector"`
	// Prometheus configuration
	PrometheusPath string `yaml:"prometheus_path"`
	// Health check configuration
	HealthCheckPort int `yaml:"health_check_port"`
	// Logging
	LogFormat string `yaml:"log_format"`
	LogOutput string `yaml:"log_output"`
}

type CircuitBreakerConfig struct {
	Enabled     bool          `yaml:"enabled"`
	MaxRequests int           `yaml:"max_requests"`
	Interval    time.Duration `yaml:"interval"`
	Timeout     time.Duration `yaml:"timeout"`
	// Rate limiting
	RateLimit       int           `yaml:"rate_limit"`
	RateLimitPeriod time.Duration `yaml:"rate_limit_period"`
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func verifyConfigPaths() {
	paths := []string{
		"./sync/config/config.yaml",
		// "../config/config.yaml",
		// "../../config/config.yaml",
		// "/etc/digital-discovery/config.yaml",
	}

	fmt.Println("Checking config file locations:")
	for _, path := range paths {
		if fileExists(path) {
			fmt.Printf("✅ Found config at: %s\n", path)
		} else {
			fmt.Printf("❌ No config at: %s\n", path)
		}
	}
}

// LoadConfig loads configuration from both file and environment variables
func LoadConfig() (*Config, error) {
	verifyConfigPaths()

	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Add debug logging to verify defaults were set
	fmt.Printf("After defaults - healthCheckPort: %v\n", v.GetInt("monitoring.healthCheckPort"))

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./sync/config")

	// Enable environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("DD")

	// Add debug logging after env vars
	fmt.Printf("After env setup - healthCheckPort: %v\n", v.GetInt("monitoring.healthCheckPort"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Add debug logging when no config file found
		fmt.Println("No config file found, using defaults")
	}

	// Add debug logging after config read
	fmt.Printf("After config read - healthCheckPort: %v\n", v.GetInt("monitoring.healthCheckPort"))

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Add debug logging after unmarshal
	fmt.Printf("Final config - healthCheckPort: %v\n", config.Monitoring.HealthCheckPort)

	return config, nil
}

func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.environment", "development")
	v.SetDefault("app.logLevel", "info")
	v.SetDefault("app.serviceName", "digital-discovery-sync")
	v.SetDefault("app.version", "1.0.0")

	// Kafka defaults
	v.SetDefault("kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("kafka.groupId", "digital-discovery-sync")
	v.SetDefault("kafka.topicPrefix", "postgres.digital_discovery.public")
	v.SetDefault("kafka.autoOffsetReset", "earliest")
	v.SetDefault("kafka.securityEnabled", false)

	// Elasticsearch defaults
	v.SetDefault("es.hosts", []string{"http://localhost:9200"})
	v.SetDefault("es.indexPrefix", "digital-discovery")
	v.SetDefault("es.maxRetries", 3)
	v.SetDefault("es.timeout", "30s")
	v.SetDefault("es.username", "")
	v.SetDefault("es.password", "")

	// Sync defaults
	v.SetDefault("sync.mode", "kafka")
	v.SetDefault("sync.kafkaConnect.enabled", false)
	v.SetDefault("sync.kafkaConnect.url", "")
	v.SetDefault("sync.kafkaConnect.name", "")
	v.SetDefault("sync.custom.enabled", false)
	v.SetDefault("sync.custom.batchSize", 100)
	v.SetDefault("sync.custom.maxRetries", 3)
	v.SetDefault("sync.custom.retryDelay", "5s")
	v.SetDefault("sync.custom.maxRetryDelay", "1h")
	v.SetDefault("sync.custom.backoffFactor", 2.0)
	v.SetDefault("sync.custom.failureQueue", "failed-syncs")
	v.SetDefault("sync.custom.conflictMode", "timestamp")

	// Monitoring defaults
	v.SetDefault("monitoring.enabled", true)
	v.SetDefault("monitoring.metricsPort", 8085)
	v.SetDefault("monitoring.tracingEnabled", true)
	v.SetDefault("monitoring.otelCollector", "localhost:4317")
	v.SetDefault("monitoring.prometheusPath", "/metrics")
	v.SetDefault("monitoring.healthCheckPort", 8082)
	v.SetDefault("monitoring.logFormat", "json")
	v.SetDefault("monitoring.logOutput", "stdout")

	// CircuitBreaker defaults
	v.SetDefault("circuitBreaker.enabled", true)
	v.SetDefault("circuitBreaker.maxRequests", 10)
	v.SetDefault("circuitBreaker.interval", "1m")
	v.SetDefault("circuitBreaker.timeout", "10s")
	v.SetDefault("circuitBreaker.rateLimit", 10)
	v.SetDefault("circuitBreaker.rateLimitPeriod", "1m")
}
