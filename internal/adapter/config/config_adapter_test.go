package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yanosea/gosl/internal/adapter/config"
	"github.com/yanosea/gosl/internal/app/port"
)

func TestNewConfigAdapter(t *testing.T) {
	adapter := config.NewConfigAdapter()

	if adapter == nil {
		t.Fatal("NewConfigAdapter returned nil")
	}
}

func TestConfigAdapter_GetConfigPath(t *testing.T) {
	adapter := config.NewConfigAdapter()

	path := adapter.GetConfigPath()

	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() = %v, want absolute path", path)
	}

	// Should contain "gosl" directory
	if filepath.Base(filepath.Dir(path)) != "gosl" {
		t.Errorf("GetConfigPath() should be in 'gosl' directory, got %v", path)
	}
}

func TestConfigAdapter_GenerateTemplate(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gosl", "config.toml")

	adapter := config.NewConfigAdapterWithPath(configPath)

	ctx := context.Background()
	err := adapter.GenerateTemplate(ctx)

	if err != nil {
		t.Fatalf("GenerateTemplate() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("GenerateTemplate() did not create config file")
	}

	// Verify file contains expected content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read generated config: %v", err)
	}

	contentStr := string(content)

	// Check for expected keys
	expectedKeys := []string{
		"slack_token",
		"workspace_id",
		"default_channel",
		"message_limit",
		"log_level",
	}

	for _, key := range expectedKeys {
		if !contains(contentStr, key) {
			t.Errorf("Generated config missing key: %s", key)
		}
	}
}

func TestConfigAdapter_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gosl", "config.toml")

	adapter := config.NewConfigAdapterWithPath(configPath)

	ctx := context.Background()

	// Create test config
	testConfig := &port.Config{
		SlackToken:     "xoxb-test-token",
		AppToken:       "xapp-test-token",
		WorkspaceID:    "T12345678",
		DefaultChannel: "general",
		MessageLimit:   10,
		LogLevel:       "info",
	}

	// Save config
	err := adapter.Save(ctx, testConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load config
	loadedConfig, err := adapter.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded config matches saved config
	if loadedConfig.SlackToken != testConfig.SlackToken {
		t.Errorf("SlackToken = %v, want %v", loadedConfig.SlackToken, testConfig.SlackToken)
	}
	if loadedConfig.WorkspaceID != testConfig.WorkspaceID {
		t.Errorf("WorkspaceID = %v, want %v", loadedConfig.WorkspaceID, testConfig.WorkspaceID)
	}
	if loadedConfig.DefaultChannel != testConfig.DefaultChannel {
		t.Errorf("DefaultChannel = %v, want %v", loadedConfig.DefaultChannel, testConfig.DefaultChannel)
	}
	if loadedConfig.MessageLimit != testConfig.MessageLimit {
		t.Errorf("MessageLimit = %v, want %v", loadedConfig.MessageLimit, testConfig.MessageLimit)
	}
	if loadedConfig.LogLevel != testConfig.LogLevel {
		t.Errorf("LogLevel = %v, want %v", loadedConfig.LogLevel, testConfig.LogLevel)
	}
}

func TestConfigAdapter_Load_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gosl", "nonexistent.toml")

	adapter := config.NewConfigAdapterWithPath(configPath)

	ctx := context.Background()
	_, err := adapter.Load(ctx)

	if err == nil {
		t.Error("Load() should return error when file does not exist")
	}
}

func TestConfigAdapter_Load_Validation(t *testing.T) {
	tests := []struct {
		name             string
		config           *port.Config
		wantMessageLimit int
	}{
		{
			name: "message_limit out of range (too low)",
			config: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   0,
				LogLevel:       "info",
			},
			wantMessageLimit: 5, // Should be corrected to default
		},
		{
			name: "message_limit out of range (too high)",
			config: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   101,
				LogLevel:       "info",
			},
			wantMessageLimit: 5, // Should be corrected to default
		},
		{
			name: "message_limit in valid range",
			config: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   10,
				LogLevel:       "info",
			},
			wantMessageLimit: 10, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "gosl", "config.toml")

			adapter := config.NewConfigAdapterWithPath(configPath)
			ctx := context.Background()

			// Save config
			err := adapter.Save(ctx, tt.config)
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Load and validate
			loaded, err := adapter.Load(ctx)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if loaded.MessageLimit != tt.wantMessageLimit {
				t.Errorf("MessageLimit = %v, want %v", loaded.MessageLimit, tt.wantMessageLimit)
			}
		})
	}
}

func TestConfigAdapter_StrictDecoding(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gosl", "config.toml")

	// Create config file with unknown key
	invalidTOML := `
slack_token = "xoxb-test"
app_token = "xapp-test"
workspace_id = "T12345"
default_channel = "general"
message_limit = 5
log_level = "info"
unknown_key = "should cause error"
`

	// Create directory
	os.MkdirAll(filepath.Dir(configPath), 0755)

	// Write invalid config
	err := os.WriteFile(configPath, []byte(invalidTOML), 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	adapter := config.NewConfigAdapterWithPath(configPath)
	ctx := context.Background()

	// Try to load - should detect unknown key
	_, err = adapter.Load(ctx)
	if err == nil {
		t.Error("Load() should return error for unknown keys with strict decoding")
	}
}

func TestConfigAdapter_LoadWithTextWrapConfig(t *testing.T) {
	tests := []struct {
		name     string
		tomlData string
		want     *port.Config
	}{
		{
			name: "text_wrap section present with custom values",
			tomlData: `
slack_token = "xoxb-test"
app_token = "xapp-test"
workspace_id = "T12345"
default_channel = "general"
message_limit = 5
log_level = "info"

[text_wrap]
enabled = false
max_line_width = 100
break_at_cjk_punctuation = false
`,
			want: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   5,
				LogLevel:       "info",
				TextWrap: port.TextWrapConfig{
					Enabled:               false,
					MaxLineWidth:          100,
					BreakAtCJKPunctuation: false,
				},
			},
		},
		{
			name: "text_wrap section missing - should use defaults",
			tomlData: `
slack_token = "xoxb-test"
app_token = "xapp-test"
workspace_id = "T12345"
default_channel = "general"
message_limit = 5
log_level = "info"
`,
			want: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   5,
				LogLevel:       "info",
				TextWrap: port.TextWrapConfig{
					Enabled:               true,
					MaxLineWidth:          0,
					BreakAtCJKPunctuation: true,
				},
			},
		},
		{
			name: "max_line_width out of range - should fallback to default",
			tomlData: `
slack_token = "xoxb-test"
app_token = "xapp-test"
workspace_id = "T12345"
default_channel = "general"
message_limit = 5
log_level = "info"

[text_wrap]
enabled = true
max_line_width = 600
break_at_cjk_punctuation = true
`,
			want: &port.Config{
				SlackToken:     "xoxb-test",
				AppToken:       "xapp-test",
				WorkspaceID:    "T12345",
				DefaultChannel: "general",
				MessageLimit:   5,
				LogLevel:       "info",
				TextWrap: port.TextWrapConfig{
					Enabled:               true,
					MaxLineWidth:          0,
					BreakAtCJKPunctuation: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "gosl", "config.toml")

			os.MkdirAll(filepath.Dir(configPath), 0755)
			err := os.WriteFile(configPath, []byte(tt.tomlData), 0600)
			if err != nil {
				t.Fatalf("Failed to write test config: %v", err)
			}

			adapter := config.NewConfigAdapterWithPath(configPath)
			ctx := context.Background()

			loaded, err := adapter.Load(ctx)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			if loaded.TextWrap.Enabled != tt.want.TextWrap.Enabled {
				t.Errorf("TextWrap.Enabled = %v, want %v", loaded.TextWrap.Enabled, tt.want.TextWrap.Enabled)
			}
			if loaded.TextWrap.MaxLineWidth != tt.want.TextWrap.MaxLineWidth {
				t.Errorf("TextWrap.MaxLineWidth = %v, want %v", loaded.TextWrap.MaxLineWidth, tt.want.TextWrap.MaxLineWidth)
			}
			if loaded.TextWrap.BreakAtCJKPunctuation != tt.want.TextWrap.BreakAtCJKPunctuation {
				t.Errorf("TextWrap.BreakAtCJKPunctuation = %v, want %v", loaded.TextWrap.BreakAtCJKPunctuation, tt.want.TextWrap.BreakAtCJKPunctuation)
			}
		})
	}
}

func TestConfigAdapter_SaveWithTextWrapConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "gosl", "config.toml")

	adapter := config.NewConfigAdapterWithPath(configPath)
	ctx := context.Background()

	testConfig := &port.Config{
		SlackToken:     "xoxb-test-token",
		AppToken:       "xapp-test-token",
		WorkspaceID:    "T12345678",
		DefaultChannel: "general",
		MessageLimit:   10,
		LogLevel:       "info",
		TextWrap: port.TextWrapConfig{
			Enabled:               false,
			MaxLineWidth:          120,
			BreakAtCJKPunctuation: false,
		},
	}

	err := adapter.Save(ctx, testConfig)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	contentStr := string(content)
	if !contains(contentStr, "[text_wrap]") {
		t.Error("Saved config should contain [text_wrap] section")
	}
	if !contains(contentStr, "enabled") {
		t.Error("Saved config should contain 'enabled' field")
	}
	if !contains(contentStr, "max_line_width") {
		t.Error("Saved config should contain 'max_line_width' field")
	}
	if !contains(contentStr, "break_at_cjk_punctuation") {
		t.Error("Saved config should contain 'break_at_cjk_punctuation' field")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
