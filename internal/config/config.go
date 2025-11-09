package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Git        GitConfig     `mapstructure:"git"`
	Kubernetes K8sConfig     `mapstructure:"kubernetes"`
	Sync       SyncConfig    `mapstructure:"sync"`
	Webhook    WebhookConfig `mapstructure:"webhook"`
}

type GitConfig struct {
	URL       string `mapstructure:"url"`
	Branch    string `mapstructure:"branch"`
	Path      string `mapstructure:"path"`
	LocalPath string `mapstructure:"localPath"`
}

type K8sConfig struct {
	Namespace  string `mapstructure:"namespace"`
	Kubeconfig string `mapstructure:"kubeconfig"`
}

type SyncConfig struct {
	Interval time.Duration `mapstructure:"interval"`
	Prune    bool          `mapstructure:"prune"`
}

type WebhookConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Secret  string `mapstructure:"secret"`
	Port    int    `mapstructure:"port"`
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetDefault("git.branch", "master")
	v.SetDefault("git.localPath", "/tmp/gitops-repo")
	v.SetDefault("kubernetes.namespace", "default")
	v.SetDefault("sync.interval", "60s")
	v.SetDefault("sync.prune", true)
	v.SetDefault("webhook.enabled", true)
	v.SetDefault("webhook.port", 8080)

	v.SetConfigName("config")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("/etc/gitops-controller")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Println("No config file found, using defaults.")
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

	if cfg.Git.URL == "" {
		return nil, fmt.Errorf("config error: 'git.url' is required. Set it in config.yaml or via GITOPS_GIT_URL")
	}

	fmt.Println("Configuration loaded successfully.")
	return &cfg, nil
}
