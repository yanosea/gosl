package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"

	"github.com/yanosea/gosl/internal/adapter/config"
	"github.com/yanosea/gosl/internal/adapter/slack"
	"github.com/yanosea/gosl/internal/adapter/tui"
	"github.com/yanosea/gosl/internal/app/service"
	"github.com/yanosea/gosl/internal/domain/cache"
	"github.com/yanosea/gosl/internal/domain/logger"
)

const (
	AppName    = "gosl"
	AppVersion = "0.1.0"

	DefaultMaxChannels   = 20
	DefaultMaxCacheMemory = 100 * 1024 * 1024
	ConfigDirPerm        = 0700
	LogDirPerm           = 0755
)

type CLIConfig struct {
	Config  string
	Version bool
}

var cliConfig CLIConfig

func main() {
	flag.StringVar(&cliConfig.Config, "config", "", "Path to config file (default: $XDG_CONFIG_HOME/gosl/config.toml)")
	flag.BoolVar(&cliConfig.Version, "version", false, "Print version information")
	flag.Parse()

	if cliConfig.Version {
		fmt.Printf("%s version %s\n", AppName, AppVersion)
		os.Exit(0)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	configPath := cliConfig.Config
	if configPath == "" {
		configPath = getConfigPath()
	}

	app, err := initializeApp(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize application: %w", err)
	}

	return app.Run()
}

type Application struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *logger.Logger
	configAdapter   *config.ConfigAdapter
	slackAdapter    *slack.SlackAdapter
	messageCache    *cache.MessageCache
	appService      *service.AppService
	eventDispatcher *tui.EventDispatcher
	program         *tea.Program
	errgroup        *errgroup.Group
}

func initializeApp(ctx context.Context) (*Application, error) {
	loggerConfig := logger.Config{
		Level:      slog.LevelInfo,
		OutputPath: getLogPath(),
		Format:     logger.FormatText,
		AddSource:  false,
	}

	appLogger, err := logger.NewLogger(loggerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	appLogger.Info(ctx, "starting gosl", "version", AppVersion)

	var configAdapter *config.ConfigAdapter
	configPath := cliConfig.Config
	if configPath != "" {
		configAdapter = config.NewConfigAdapterWithPath(configPath)
	} else {
		configAdapter = config.NewConfigAdapter()
	}

	cfg, err := configAdapter.Load(ctx)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			configPath := configAdapter.GetConfigPath()
			appLogger.Info(ctx, "config file not found, generating template", "path", configPath)
			if err := configAdapter.GenerateTemplate(ctx); err != nil {
				return nil, fmt.Errorf("failed to generate config template: %w", err)
			}
			fmt.Printf("configuration template created at: %s\n", configPath)
			fmt.Println("please edit the configuration file and add your Slack token, then restart gosl.")
			os.Exit(0)
		}
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	appLogger.Info(ctx, "configuration loaded", "message_limit", cfg.MessageLimit)

	slackAdapter := slack.NewSlackAdapter()
	messageCache := cache.NewMessageCache(DefaultMaxChannels, DefaultMaxCacheMemory)
	appService := service.NewAppService(configAdapter, slackAdapter, messageCache)
	appModel := tui.NewAppModel(appService, cfg)

	program := tea.NewProgram(
		appModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	eventDispatcher := tui.NewEventDispatcher()

	appCtx, appCancel := context.WithCancel(ctx)
	g, gCtx := errgroup.WithContext(appCtx)

	app := &Application{
		ctx:             appCtx,
		cancel:          appCancel,
		logger:          appLogger,
		configAdapter:   configAdapter,
		slackAdapter:    slackAdapter,
		messageCache:    messageCache,
		appService:      appService,
		eventDispatcher: eventDispatcher,
		program:         program,
		errgroup:        g,
	}

	g.Go(func() error {
		if err := slackAdapter.Connect(gCtx, cfg.SlackToken, cfg.AppToken); err != nil {
			appLogger.Error(gCtx, "failed to connect to Slack", err)
			program.Send(tui.ErrorMsg{Err: fmt.Sprintf("failed to connect to Slack: %v", err)})
			return fmt.Errorf("failed to connect to Slack: %w", err)
		}

		appLogger.Info(gCtx, "connected to Slack")

		eventChan, err := slackAdapter.SubscribeEvents(gCtx)
		if err != nil {
			appLogger.Error(gCtx, "failed to subscribe to Slack events", err)
			return fmt.Errorf("failed to subscribe to Slack events: %w", err)
		}

		if err := eventDispatcher.Start(gCtx, eventChan, program); err != nil {
			appLogger.Error(gCtx, "failed to start event dispatcher", err)
			return fmt.Errorf("failed to start event dispatcher: %w", err)
		}

		return nil
	})

	return app, nil
}

func (a *Application) Run() error {
	defer func() {
		a.cancel()
		a.eventDispatcher.Stop()
		if err := a.slackAdapter.Disconnect(); err != nil {
			a.logger.Error(a.ctx, "failed to disconnect from Slack", err)
		}
		a.logger.Info(a.ctx, "application shutdown complete")
	}()

	if _, err := a.program.Run(); err != nil {
		return fmt.Errorf("tui error: %w", err)
	}

	if err := a.errgroup.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func getConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return filepath.Join(".", "config.toml")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "gosl", "config.toml")
}

func getLogPath() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return filepath.Join(".", "gosl.log")
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "gosl", "logs", "gosl.log")
}
