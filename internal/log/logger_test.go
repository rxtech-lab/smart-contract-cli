package log

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LoggerTestSuite struct {
	suite.Suite
	tempDir string
}

func (s *LoggerTestSuite) SetupTest() {
	// Create a temporary directory for test logs
	tempDir, err := os.MkdirTemp("", "logger-test-*")
	s.Require().NoError(err)
	s.tempDir = tempDir
}

func (s *LoggerTestSuite) TearDownTest() {
	// Clean up temporary directory
	if s.tempDir != "" {
		_ = os.RemoveAll(s.tempDir)
	}
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) TestNewLogger() {
	logger := NewLogger()
	s.NotNil(logger)
	s.NotNil(logger.Logger)
	s.Nil(logger.file)
}

func (s *LoggerTestSuite) TestDefaultConfig() {
	config := DefaultConfig()
	s.Equal("./logs/app.log", config.LogFilePath)
	s.Equal(10, config.MaxSize)
	s.Equal(5, config.MaxBackups)
	s.Equal(30, config.MaxAge)
	s.True(config.Compress)
	s.True(config.ConsoleOutput)
}

func (s *LoggerTestSuite) TestNewLoggerWithConfig() {
	logPath := filepath.Join(s.tempDir, "test.log")
	config := Config{
		LogFilePath:   logPath,
		MaxSize:       5,
		MaxBackups:    3,
		MaxAge:        7,
		Compress:      false,
		ConsoleOutput: false,
	}

	logger, err := NewLoggerWithConfig(config)
	s.Require().NoError(err)
	s.NotNil(logger)
	s.NotNil(logger.Logger)
	s.NotNil(logger.file)

	// Clean up
	defer func() { _ = logger.Close() }()

	// Verify log directory was created
	logDir := filepath.Dir(logPath)
	_, err = os.Stat(logDir)
	s.NoError(err)
}

func (s *LoggerTestSuite) TestNewFileLogger() {
	logPath := filepath.Join(s.tempDir, "file-only.log")

	logger, err := NewFileLogger(logPath)
	s.Require().NoError(err)
	s.NotNil(logger)
	s.NotNil(logger.file)

	defer func() { _ = logger.Close() }()
}

func (s *LoggerTestSuite) TestWriteToFile() {
	logPath := filepath.Join(s.tempDir, "write-test.log")
	config := Config{
		LogFilePath:   logPath,
		MaxSize:       10,
		MaxBackups:    5,
		MaxAge:        30,
		Compress:      false,
		ConsoleOutput: false,
	}

	logger, err := NewLoggerWithConfig(config)
	s.Require().NoError(err)
	defer func() { _ = logger.Close() }()

	// Write some logs
	logger.Info("Info message: %s", "test")
	logger.Error("Error message: %d", 123)
	logger.Debug("Debug message")
	logger.Warn("Warning message")

	// Close to flush
	err = logger.Close()
	s.NoError(err)

	// Read the log file
	content, err := os.ReadFile(logPath) // #nosec G304 - test file path is controlled
	s.Require().NoError(err)

	logContent := string(content)
	s.Contains(logContent, "Info message: test")
	s.Contains(logContent, "Error message: 123")
	s.Contains(logContent, "Debug message")
	s.Contains(logContent, "Warning message")
	s.Contains(logContent, `"level":"info"`)
	s.Contains(logContent, `"level":"error"`)
	s.Contains(logContent, `"level":"debug"`)
	s.Contains(logContent, `"level":"warn"`)
}

func (s *LoggerTestSuite) TestLogLevels() {
	logPath := filepath.Join(s.tempDir, "levels.log")
	logger, err := NewFileLogger(logPath)
	s.Require().NoError(err)
	defer func() { _ = logger.Close() }()

	logger.Info("info")
	logger.Error("error")
	logger.Debug("debug")
	logger.Warn("warn")

	_ = logger.Close()

	content, err := os.ReadFile(logPath) // #nosec G304 - test file path is controlled
	s.Require().NoError(err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	s.Len(lines, 4)
}

func (s *LoggerTestSuite) TestRotate() {
	logPath := filepath.Join(s.tempDir, "rotate.log")
	logger, err := NewFileLogger(logPath)
	s.Require().NoError(err)
	defer func() { _ = logger.Close() }()

	logger.Info("Before rotation")

	// Trigger rotation
	err = logger.Rotate()
	s.NoError(err)

	logger.Info("After rotation")
	_ = logger.Close()

	// Check that log file exists
	_, err = os.Stat(logPath)
	s.NoError(err)
}

func (s *LoggerTestSuite) TestCloseNilFile() {
	logger := NewLogger()
	err := logger.Close()
	s.NoError(err)
}

func (s *LoggerTestSuite) TestRotateNilFile() {
	logger := NewLogger()
	err := logger.Rotate()
	s.NoError(err)
}

func (s *LoggerTestSuite) TestLogDirectoryCreation() {
	logPath := filepath.Join(s.tempDir, "nested", "deep", "path", "test.log")
	config := DefaultConfig()
	config.LogFilePath = logPath
	config.ConsoleOutput = false

	logger, err := NewLoggerWithConfig(config)
	s.Require().NoError(err)
	defer func() { _ = logger.Close() }()

	logger.Info("Test message")
	_ = logger.Close()

	// Verify the nested directory structure was created
	_, err = os.Stat(filepath.Dir(logPath))
	s.NoError(err)

	// Verify log file exists
	_, err = os.Stat(logPath)
	s.NoError(err)
}

func (s *LoggerTestSuite) TestMultipleLoggers() {
	logPath1 := filepath.Join(s.tempDir, "logger1.log")
	logPath2 := filepath.Join(s.tempDir, "logger2.log")

	logger1, err := NewFileLogger(logPath1)
	s.Require().NoError(err)
	defer func() { _ = logger1.Close() }()

	logger2, err := NewFileLogger(logPath2)
	s.Require().NoError(err)
	defer func() { _ = logger2.Close() }()

	logger1.Info("Logger 1 message")
	logger2.Info("Logger 2 message")

	_ = logger1.Close()
	_ = logger2.Close()

	// Verify both log files exist and have correct content
	content1, err := os.ReadFile(logPath1) // #nosec G304 - test file path is controlled
	s.Require().NoError(err)
	s.Contains(string(content1), "Logger 1 message")
	s.NotContains(string(content1), "Logger 2 message")

	content2, err := os.ReadFile(logPath2) // #nosec G304 - test file path is controlled
	s.Require().NoError(err)
	s.Contains(string(content2), "Logger 2 message")
	s.NotContains(string(content2), "Logger 1 message")
}

func (s *LoggerTestSuite) TestLogWithFormatting() {
	logPath := filepath.Join(s.tempDir, "formatting.log")
	logger, err := NewFileLogger(logPath)
	s.Require().NoError(err)
	defer func() { _ = logger.Close() }()

	logger.Info("User %s logged in with ID %d", "john_doe", 12345)
	logger.Error("Failed to process transaction %s: %s", "TX123", "insufficient funds")

	_ = logger.Close()

	content, err := os.ReadFile(logPath) // #nosec G304 - test file path is controlled
	s.Require().NoError(err)

	logContent := string(content)
	s.Contains(logContent, "User john_doe logged in with ID 12345")
	s.Contains(logContent, "Failed to process transaction TX123: insufficient funds")
}
