package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/yanosea/gosl/internal/app/port"
)

var _ port.ConfigRepository = (*ConfigAdapter)(nil)

const (
	DefaultMessageLimit  = 5
	MinMessageLimit      = 1
	MaxMessageLimit      = 100
	ConfigFilePermission = 0600
	ConfigDirPermission  = 0700
)

var (
	ErrInvalidConfig = errors.New("invalid configuration")
	ErrUnknownKeys   = errors.New("unknown keys in configuration file")

	validLogLevels = map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
)

type ConfigAdapter struct {
	configPath string
}

func NewConfigAdapter() *ConfigAdapter {
	configPath := getDefaultConfigPath()
	return &ConfigAdapter{
		configPath: configPath,
	}
}

func NewConfigAdapterWithPath(path string) *ConfigAdapter {
	return &ConfigAdapter{
		configPath: path,
	}
}

func getDefaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			configHome = filepath.Join(".", ".config")
		} else {
			configHome = filepath.Join(homeDir, ".config")
		}
	}

	return filepath.Join(configHome, "gosl", "config.toml")
}

func (a *ConfigAdapter) GetConfigPath() string {
	return a.configPath
}

func (a *ConfigAdapter) Load(ctx context.Context) (*port.Config, error) {
	var config port.Config

	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	md, err := toml.Decode(string(data), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse toml: %w", err)
	}

	undecoded := md.Undecoded()
	if len(undecoded) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrUnknownKeys, undecoded)
	}

	if err := a.validateAndCorrect(&config); err != nil {
		return nil, err
	}

	a.applyTextWrapDefaults(&config)

	return &config, nil
}

func (a *ConfigAdapter) Save(ctx context.Context, config *port.Config) error {
	if err := a.validateAndCorrect(config); err != nil {
		return err
	}

	dir := filepath.Dir(a.configPath)
	if err := os.MkdirAll(dir, ConfigDirPermission); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.OpenFile(a.configPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, ConfigFilePermission)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config to toml: %w", err)
	}

	return nil
}

func (a *ConfigAdapter) GenerateTemplate(ctx context.Context) error {
	template := &port.Config{
		SlackToken:     "xoxb-YOUR-BOT-TOKEN-HERE",
		AppToken:       "xapp-YOUR-APP-LEVEL-TOKEN-HERE",
		WorkspaceID:    "T0000000000",
		DefaultChannel: "general",
		MessageLimit:   DefaultMessageLimit,
		LogLevel:       "info",
		TextWrap:       port.DefaultTextWrapConfig(),
	}

	return a.Save(ctx, template)
}

func (a *ConfigAdapter) validateAndCorrect(config *port.Config) error {
	if config.SlackToken == "" {
		return fmt.Errorf("%w: slack_token cannot be empty", ErrInvalidConfig)
	}
	if !strings.HasPrefix(config.SlackToken, "xoxb-") && !strings.HasPrefix(config.SlackToken, "xapp-") {
		return fmt.Errorf("%w: slack_token must start with 'xoxb-' or 'xapp-'", ErrInvalidConfig)
	}

	if config.AppToken == "" {
		return fmt.Errorf("%w: app_token cannot be empty", ErrInvalidConfig)
	}
	if !strings.HasPrefix(config.AppToken, "xapp-") {
		return fmt.Errorf("%w: app_token must start with 'xapp-'", ErrInvalidConfig)
	}

	if config.LogLevel == "" {
		config.LogLevel = "info"
	} else {
		normalizedLevel := strings.ToLower(config.LogLevel)
		if !validLogLevels[normalizedLevel] {
			return fmt.Errorf("%w: log_level must be one of: debug, info, warn, error", ErrInvalidConfig)
		}
		config.LogLevel = normalizedLevel
	}

	if config.MessageLimit < MinMessageLimit || config.MessageLimit > MaxMessageLimit {
		config.MessageLimit = DefaultMessageLimit
	}

	return nil
}

func (a *ConfigAdapter) applyTextWrapDefaults(config *port.Config) {
	if config.TextWrap == (port.TextWrapConfig{}) {
		config.TextWrap = port.DefaultTextWrapConfig()
		return
	}

	if config.TextWrap.MaxLineWidth < 0 || (config.TextWrap.MaxLineWidth > 0 && (config.TextWrap.MaxLineWidth < 20 || config.TextWrap.MaxLineWidth > 500)) {
		config.TextWrap.MaxLineWidth = 0
	}
}
