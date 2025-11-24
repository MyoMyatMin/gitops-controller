package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/MyoMyatMin/gitops-controller/internal/log"
	"github.com/spf13/viper"
)

type Config struct {
	Kubernetes   K8sConfig          `mapstructure:"kubernetes"`
	Webhook      WebhookConfig      `mapstructure:"webhook"`
	Repositories []RepositoryConfig `mapstructure:"repositories"`
}
type RepositoryConfig struct {
	Name      string        `mapstructure:"name"`
	URL       string        `mapstructure:"url"`
	Branch    string        `mapstructure:"branch"`
	Path      string        `mapstructure:"path"`
	Namespace string        `mapstructure:"namespace"`
	Interval  time.Duration `mapstructure:"interval"`
	Prune     bool          `mapstructure:"prune"`
}

type K8sConfig struct {
	Kubeconfig string `mapstructure:"kubeconfig"`
}

type WebhookConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Secret  string `mapstructure:"secret"`
	Port    int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("webhook.enabled", true)
	v.SetDefault("webhook.port", 8080)

	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("No config file found, using defaults.")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	v.SetEnvPrefix("GITOPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if len(cfg.Repositories) == 0 {
		return nil, fmt.Errorf("config error: no 'repositories' defined")
	}

	log.Info("Configuration loaded successfully.")
	return &cfg, nil
}
