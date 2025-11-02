# Logger Usage Guide

This guide shows how to use the logger in your application to write logs to local files.

## Quick Start

### Option 1: Console Only (Default)

For backward compatibility, the default logger writes only to console:

```go
package main

import "github.com/rxtech-lab/smart-contract-cli/internal/log"

func main() {
    logger := log.NewLogger()
    logger.Info("Application started")
    logger.Error("An error occurred: %s", "connection failed")
}
```

### Option 2: File Only

Write logs only to a file (no console output):

```go
package main

import (
    "github.com/rxtech-lab/smart-contract-cli/internal/log"
)

func main() {
    logger, err := log.NewFileLogger("./logs/app.log")
    if err != nil {
        panic(err)
    }
    defer logger.Close()

    logger.Info("Application started")
    logger.Error("An error occurred")
}
```

### Option 3: Both Console and File (Recommended)

Write logs to both console and file with default configuration:

```go
package main

import (
    "github.com/rxtech-lab/smart-contract-cli/internal/log"
)

func main() {
    config := log.DefaultConfig()
    logger, err := log.NewLoggerWithConfig(config)
    if err != nil {
        panic(err)
    }
    defer logger.Close()

    logger.Info("Application started")
    logger.Warn("Warning: disk space running low")
    logger.Error("Error: %s", "failed to connect")
}
```

### Option 4: Custom Configuration

Full control over logging behavior:

```go
package main

import (
    "github.com/rxtech-lab/smart-contract-cli/internal/log"
)

func main() {
    config := log.Config{
        LogFilePath:   "./logs/myapp.log",
        MaxSize:       20,    // MB
        MaxBackups:    10,    // files
        MaxAge:        60,    // days
        Compress:      true,  // gzip old files
        ConsoleOutput: true,  // also write to console
    }

    logger, err := log.NewLoggerWithConfig(config)
    if err != nil {
        panic(err)
    }
    defer logger.Close()

    logger.Info("Custom logger initialized")
}
```

## Using in app/page.go

Here's how to integrate the logger into your Bubble Tea application:

### Step 1: Add logger to your Model

```go
package app

import (
    "github.com/rxtech-lab/smart-contract-cli/internal/log"
    "github.com/rxtech-lab/smart-contract-cli/internal/view"
    // ... other imports
)

type Model struct {
    router       view.Router
    sharedMemory storage.SharedMemory
    logger       *log.Logger  // Add logger field

    // ... other fields
}
```

### Step 2: Initialize logger in NewPage

```go
func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
    // Initialize logger with file output
    logger, err := log.NewLoggerWithConfig(log.DefaultConfig())
    if err != nil {
        // Fallback to console-only logger
        logger = log.NewLogger()
    }

    model := Model{
        router:       router,
        sharedMemory: sharedMemory,
        logger:       logger,
        // ... other fields
    }

    logger.Info("Page initialized")

    return model
}
```

### Step 3: Use logger throughout your code

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.logger.Debug("Update called with message type: %T", msg)

    if !m.isUnlocked {
        return m.handlePasswordInput(msg)
    }

    keyMsg, ok := msg.(tea.KeyMsg)
    if !ok {
        return m, nil
    }

    m.logger.Info("Key pressed: %s", keyMsg.String())

    switch keyMsg.String() {
    case "enter", " ":
        m.logger.Info("Navigating to route: %s", m.selectedOption.Route)
        err := m.router.NavigateTo(m.selectedOption.Route, nil)
        if err != nil {
            m.logger.Error("Navigation failed: %v", err)
            return m, tea.Quit
        }
    case "q", "ctrl+c":
        m.logger.Info("User quit application")
        return m, tea.Quit
    }

    return m, nil
}
```

### Step 4: Log errors and important events

```go
func (m Model) handlePasswordSubmit(password string) (Model, tea.Cmd) {
    if password == "" {
        m.logger.Warn("Empty password submitted")
        m.errorMessage = "Password cannot be empty"
        return m, nil
    }

    m.logger.Info("Attempting to unlock storage")

    if err := m.ensureSecureStorageInitialized(); err != nil {
        m.logger.Error("Failed to initialize storage: %v", err)
        m.errorMessage = fmt.Sprintf("Failed to initialize storage: %v", err)
        return m, nil
    }

    if err := m.unlockAndStorePassword(password); err != nil {
        m.logger.Error("Failed to unlock: %v", err)
        m.errorMessage = fmt.Sprintf("Failed to unlock: %v", err)
        return m, nil
    }

    m.logger.Info("Storage unlocked successfully")
    m.isUnlocked = true
    m.errorMessage = ""
    return m, nil
}
```

### Step 5: Clean up (optional)

If you want to ensure logs are flushed when the app exits, you can add cleanup:

```go
// In your main.go or wherever you start the Bubble Tea program
func main() {
    // Initialize logger
    logger, err := log.NewLoggerWithConfig(log.DefaultConfig())
    if err != nil {
        panic(err)
    }
    defer logger.Close()  // Ensure logs are flushed on exit

    // Create your model with the logger
    model := app.NewPage(router, sharedMemory)

    // Run the program
    p := tea.NewProgram(model)
    if _, err := p.Run(); err != nil {
        logger.Fatal("Application error: %v", err)
    }
}
```

## Log Levels

The logger supports different log levels:

```go
logger.Debug("Detailed debugging information")  // Use for development
logger.Info("General information")              // Normal application flow
logger.Warn("Warning: something unexpected")    // Warnings
logger.Error("Error occurred: %v", err)        // Errors
logger.Fatal("Fatal error: %v", err)           // Fatal errors (exits app)
```

## Log Rotation

Logs are automatically rotated based on configuration:

- **MaxSize**: When log file reaches this size (MB), it rotates
- **MaxBackups**: Number of old log files to keep
- **MaxAge**: Days to keep old log files
- **Compress**: Automatically gzip old log files

Example log file structure after rotation:
```
logs/
├── app.log           (current log)
├── app.log.1         (most recent backup)
├── app.log.2.gz      (older, compressed)
├── app.log.3.gz
└── app.log.4.gz
```

You can also manually trigger rotation:

```go
err := logger.Rotate()
if err != nil {
    logger.Error("Failed to rotate logs: %v", err)
}
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| LogFilePath | string | "./logs/app.log" | Path to log file |
| MaxSize | int | 10 | Max size in MB before rotation |
| MaxBackups | int | 5 | Max number of old log files |
| MaxAge | int | 30 | Max days to keep old files |
| Compress | bool | true | Compress rotated files |
| ConsoleOutput | bool | true | Also write to console |

## Best Practices

1. **Always close the logger** when using file output:
   ```go
   defer logger.Close()
   ```

2. **Use appropriate log levels**:
   - Debug: Development and troubleshooting
   - Info: Normal application flow
   - Warn: Unexpected but handled situations
   - Error: Errors that need attention
   - Fatal: Unrecoverable errors

3. **Include context in log messages**:
   ```go
   logger.Info("User %s logged in from %s", username, ipAddress)
   logger.Error("Database query failed: query=%s, error=%v", query, err)
   ```

4. **Don't log sensitive information**:
   ```go
   // BAD
   logger.Info("Password: %s", password)

   // GOOD
   logger.Info("User authenticated successfully")
   ```

5. **Use structured logging for complex data**:
   ```go
   logger.Info("Transaction completed: id=%s, amount=%d, user=%s",
       txID, amount, userID)
   ```

## Example: Full Integration

Here's a complete example integrating the logger into your app:

```go
package app

import (
    "fmt"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/rxtech-lab/smart-contract-cli/internal/log"
    "github.com/rxtech-lab/smart-contract-cli/internal/storage"
    "github.com/rxtech-lab/smart-contract-cli/internal/view"
)

type Model struct {
    router       view.Router
    sharedMemory storage.SharedMemory
    logger       *log.Logger
    // ... other fields
}

func NewPage(router view.Router, sharedMemory storage.SharedMemory) view.View {
    // Create logger with custom config
    config := log.Config{
        LogFilePath:   "./logs/smart-contract-cli.log",
        MaxSize:       20,  // 20 MB
        MaxBackups:    7,   // Keep 7 backups
        MaxAge:        90,  // Keep for 90 days
        Compress:      true,
        ConsoleOutput: false, // File only in production
    }

    logger, err := log.NewLoggerWithConfig(config)
    if err != nil {
        // Fallback to console logger
        logger = log.NewLogger()
        logger.Warn("Failed to create file logger, using console: %v", err)
    }

    logger.Info("Application started")

    model := Model{
        router:       router,
        sharedMemory: sharedMemory,
        logger:       logger,
    }

    return model
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.logger.Debug("Processing message: %T", msg)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        m.logger.Info("Key event: %s", msg.String())

        switch msg.String() {
        case "q", "ctrl+c":
            m.logger.Info("User requested quit")
            m.logger.Close() // Clean shutdown
            return m, tea.Quit
        }
    }

    return m, nil
}
```

## Troubleshooting

### Logs not appearing in file

1. Check file permissions on the log directory
2. Ensure `logger.Close()` is called to flush pending writes
3. Verify the log file path is correct

### Log rotation not working

1. Check `MaxSize` configuration
2. Ensure the application has write permissions
3. Verify disk space is available

### Performance issues

1. Reduce log level (use Info/Warn/Error only, not Debug)
2. Increase `MaxSize` to reduce rotation frequency
3. Set `Compress: false` if CPU is constrained
4. Set `ConsoleOutput: false` in production
