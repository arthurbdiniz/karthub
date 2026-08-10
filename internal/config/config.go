package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Session  SessionConfig  `yaml:"session"`
	Mail     MailConfig     `yaml:"mail"`
	Push     PushConfig     `yaml:"push"`
	Log      LogConfig      `yaml:"log"`
	BaseURL  string         `yaml:"base_url"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type SessionConfig struct {
	Secret   string `yaml:"secret"`
	MaxAge   int    `yaml:"max_age"`
	Secure   bool   `yaml:"secure"`
	HTTPOnly bool   `yaml:"http_only"`
}

type MailConfig struct {
	ResendAPIKey string `yaml:"resend_api_key"`
	FromAddress  string `yaml:"from_address"`
	FromName     string `yaml:"from_name"`
	DevMode      bool   `yaml:"dev_mode"`
}

type PushConfig struct {
	VAPIDPublicKey  string `yaml:"vapid_public_key"`
	VAPIDPrivateKey string `yaml:"vapid_private_key"`
	Subscriber      string `yaml:"subscriber"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func Load() (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile("config.yaml")
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if data != nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parsing config: %w", err)
		}
	}

	applyEnvOverrides(cfg)
	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Path: "data/karthub.db",
		},
		Session: SessionConfig{
			Secret:   "change-me-in-production",
			MaxAge:   86400 * 30, // 30 days
			Secure:   false,
			HTTPOnly: true,
		},
		Mail: MailConfig{
			FromAddress: "noreply@example.com",
			FromName:    "KartHub",
			DevMode:     true,
		},
		Log: LogConfig{
			Level: "info",
		},
		BaseURL: "http://localhost:8080",
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("KARTHUB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Server.Port = port
		}
	}
	if v := os.Getenv("KARTHUB_DB_PATH"); v != "" {
		cfg.Database.Path = v
	}
	if v := os.Getenv("KARTHUB_SESSION_SECRET"); v != "" {
		cfg.Session.Secret = v
	}
	if v := os.Getenv("KARTHUB_LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("KARTHUB_RESEND_API_KEY"); v != "" {
		cfg.Mail.ResendAPIKey = v
		cfg.Mail.DevMode = false // auto-disable dev mode when API key is set
	}
	if v := os.Getenv("KARTHUB_MAIL_FROM"); v != "" {
		cfg.Mail.FromAddress = v
	}
	if v := os.Getenv("KARTHUB_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v := os.Getenv("KARTHUB_VAPID_PUBLIC_KEY"); v != "" {
		cfg.Push.VAPIDPublicKey = v
	}
	if v := os.Getenv("KARTHUB_VAPID_PRIVATE_KEY"); v != "" {
		cfg.Push.VAPIDPrivateKey = v
	}
	if v := os.Getenv("KARTHUB_VAPID_SUBSCRIBER"); v != "" {
		cfg.Push.Subscriber = v
	}
}
