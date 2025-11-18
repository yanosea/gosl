package port

import (
	"context"

	"github.com/yanosea/gosl/internal/domain/textwrap"
)

type Config struct {
	SlackToken     string             `toml:"slack_token"`
	AppToken       string             `toml:"app_token"`
	WorkspaceID    string             `toml:"workspace_id"`
	DefaultChannel string             `toml:"default_channel"`
	MessageLimit   int                `toml:"message_limit"`
	LogLevel       string             `toml:"log_level"`
	TextWrap       TextWrapConfig     `toml:"text_wrap"`
}

type TextWrapConfig struct {
	Enabled               bool `toml:"enabled"`
	MaxLineWidth          int  `toml:"max_line_width"`
	BreakAtCJKPunctuation bool `toml:"break_at_cjk_punctuation"`
}

func DefaultTextWrapConfig() TextWrapConfig {
	return TextWrapConfig{
		Enabled:               true,
		MaxLineWidth:          0,
		BreakAtCJKPunctuation: true,
	}
}

func (c TextWrapConfig) ToOptions() textwrap.TextWrapOptions {
	return textwrap.TextWrapOptions{
		Enabled:               c.Enabled,
		MaxLineWidth:          c.MaxLineWidth,
		BreakAtCJKPunctuation: c.BreakAtCJKPunctuation,
	}
}

type ConfigRepository interface {
	Load(ctx context.Context) (*Config, error)
	Save(ctx context.Context, config *Config) error
	GenerateTemplate(ctx context.Context) error
	GetConfigPath() string
}
