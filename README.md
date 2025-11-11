# ⚠️ CAUTION ⚠️

This repository is fully generated with Claude Code Spec Drriven Development.

<div align="right">

![golangci-lint](https://github.com/yanosea/gosl/actions/workflows/golangci-lint.yml/badge.svg)
![release](https://github.com/yanosea/gosl/actions/workflows/release.yml/badge.svg)

</div>

<div align="center">

# 💬 gosl

![Language:Go](https://img.shields.io/static/v1?label=Language&message=Go&color=blue&style=flat-square)
![License:MIT](https://img.shields.io/static/v1?label=License&message=MIT&color=blue&style=flat-square)
[![Latest Release](https://img.shields.io/github/v/release/yanosea/gosl?style=flat-square)](https://github.com/yanosea/gosl/releases/latest)
<br/>
[Coverage Report](https://yanosea.github.io/gosl/coverage.html)

</div>

## ℹ️ About

`gosl` is a terminal-based Slack client with a rich text user interface for interacting with Slack workspaces directly from the command line.

## ✨ Features

- **🗂️ Channel Navigation** - Browse and switch between public channels, private channels, and direct messages
- **💬 Real-time Messaging** - View, send, and receive Slack messages with live updates via WebSocket
- **🧵 Thread Conversations** - Read and participate in threaded discussions
- **⌨️ Keyboard-driven UI** - Fully interactive TUI with Vim-style keyboard shortcuts
- **💾 Message Caching** - Intelligent LRU caching system for improved performance and reduced API calls
- **🪶 Lightweight** - Fast, efficient resource usage without a heavy desktop app

## 🎯 Use Cases

- **👨‍💻 Terminal-focused developers** - Stay in your terminal environment
- **🚀 Lightweight Slack access** - Quick interactions without opening a desktop app
- **🌐 Remote/SSH environments** - Access Slack from headless or remote systems
- **⚡ Keyboard-driven workflows** - Navigate Slack efficiently with keyboard shortcuts

## 💻 Usage

### ⌨️ Keyboard Shortcuts

#### Global

- `q` / `Ctrl+C` - Quit
- `?` - Show help

#### Channel List

- `↑` / `k` - Move up
- `↓` / `j` - Move down
- `Enter` - Select channel
- `/` - Search channels
- `Esc` - Exit search

#### Message View

- `↑` / `k` - Previous message
- `↓` / `j` - Next message
- `g` - Jump to top
- `G` - Jump to bottom
- `Ctrl+U` / `PgUp` - Page up
- `Ctrl+D` / `PgDn` - Page down
- `Enter` - Open thread (if exists)
- `i` / `c` - Compose message
- `Esc` - Back to channel list

#### Thread View

- `↑` / `k` - Previous message
- `↓` / `j` - Next message
- `i` / `r` - Reply to thread
- `Esc` - Back to message view

#### Input Mode

- `Ctrl+Enter` / `Alt+Enter` / `Ctrl+J` - Send message
- `Enter` - New line
- `Esc` - Cancel

### 🌍 Environments

#### 📁 Configuration file path

Default : `$XDG_CONFIG_HOME/gosl/config.toml` or `$HOME/.config/gosl/config.toml`

```sh
# Run with custom config path
gosl --config /path/to/your/config.toml
```

#### 📁 Log file path

Default : `$XDG_DATA_HOME/gosl/logs/gosl.log` or `$HOME/.local/share/gosl/logs/gosl.log`

Log path is determined by XDG Base Directory specification.

### 🔧 Installation

#### 🐭 Using go

```sh
go install github.com/yanosea/gosl/cmd/gosl@latest
```

#### 🍺 Using homebrew

```sh
brew tap yanosea/tap
brew install yanosea/tap/gosl
```

#### 📦 Download from release

Go to the [Releases](https://github.com/yanosea/gosl/releases) and download the latest binary for your platform.

#### 🚀 Build from source

```sh
git clone https://github.com/yanosea/gosl.git
cd gosl

# Build for current platform
go build -o gosl ./cmd/gosl

# Or build for specific platform
make build.linux     # Linux AMD64
make build.darwin    # macOS (Intel & Apple Silicon)
make build.windows   # Windows AMD64
```

Cross-platform binaries will be output to `./bin/`

### ⚙️ Setup

#### 1️⃣ Create a Slack App

1. Go to [Slack API](https://api.slack.com/apps)
2. Create a new app (from scratch)
3. Add Bot Token Scopes:
   - `channels:history`
   - `channels:read`
   - `chat:write`
   - `groups:history`
   - `groups:read`
   - `im:history`
   - `im:read`
   - `users:read`
4. Enable Socket Mode
5. Generate an App-Level Token with `connections:write` scope
6. Install the app to your workspace

#### 2️⃣ Configure gosl

Run gosl for the first time to generate a configuration template:

```sh
gosl
```

This creates `~/.config/gosl/config.toml`. Edit it and add your tokens:

```toml
slack_token = "xoxb-your-bot-token"
app_token = "xapp-your-app-token"
message_limit = 100
```

#### 3️⃣ Run

```sh
gosl
```

Optional flags:

- `--config <path>` - Custom config file path
- `--version` - Print version information

### ✨ Update

#### 🐭 Using go

Reinstall `gosl`!

```sh
go install github.com/yanosea/gosl/cmd/gosl@latest
```

#### 🍺 Using homebrew

```sh
brew update
brew upgrade gosl
```

#### 📦 Download from release

Download the latest binary from the [Releases](https://github.com/yanosea/gosl/releases) page and replace the old binary in your `$PATH`.

### 🧹 Uninstallation

#### 🔧 Uninstall gosl

##### 🐭 Using go

```sh
rm $GOPATH/bin/gosl
# maybe you have to execute with sudo
rm -fr $GOPATH/pkg/mod/github.com/yanosea/gosl*
```

##### 🍺 Using homebrew

```sh
brew uninstall gosl
brew untap yanosea/tap/gosl
```

##### 📦 Download from release

Remove the binary you downloaded and placed in your `$PATH`.

#### 🗑️ Remove data files

If you've set custom paths, please replace `$HOME/.config/gosl` and `$HOME/.local/share/gosl` with paths you've set.
These below commands are in the case of default. Of course you can remove whole the directory.

##### 💾 Remove configuration file

```sh
rm $HOME/.config/gosl/config.toml
```

##### 💾 Remove log files

```sh
rm -rf $HOME/.local/share/gosl/logs
```

## 🛠️ Development

### 🧪 Run Tests

```sh
# All tests with coverage
make test

# Specific package
go test ./internal/domain/...

# Unit tests only
go test -short ./...
```

### 🔄 Update Mocks

```sh
make update.mocks
```

### 📜 Update Credits

```sh
make update.credits
```

### 🧹 Clean

```sh
make clean
```

## 📃 License

[🔓MIT](./LICENSE)

## 🖊️ Author

[🏹 yanosea](https://github.com/yanosea)

## 🤝 Contributing

Feel free to point me in the right direction🙏

## 🙏 Credits

Built with amazing open-source libraries. See [CREDITS](CREDITS) for full list.
