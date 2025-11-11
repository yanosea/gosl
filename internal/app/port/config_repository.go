package port

import "context"

type Config struct {
	SlackToken     string `toml:"slack_token"`
	AppToken       string `toml:"app_token"`
	WorkspaceID    string `toml:"workspace_id"`
	DefaultChannel string `toml:"default_channel"`
	MessageLimit   int    `toml:"message_limit"`
	LogLevel       string `toml:"log_level"`
}

type ConfigRepository interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, config *Config) error
	GenerateTemplate(ctx context.Context) error
	GetConfigPath() string
}
