package logging

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type LogLevel string

const (
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

type Logger struct {
	commonLog *log.Logger
	errorLog  *log.Logger
}

// Global logger instance
var globalLogger *Logger

// This function initializes the global logger instance
func InitLogger(logDir string) error {
	logger, err := newLogger(logDir)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// This function returns the global logger instance
func GetLogger() *Logger {
	if globalLogger == nil {
		panic("Logger not initialized. Call InitLogger first")
	}
	return globalLogger
}

// This function creates a new logger instance
func newLogger(logDir string) (*Logger, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	commonLogFile, err := os.OpenFile(
		filepath.Join(logDir, "common.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open common log file: %v", err)
	}

	errorLogFile, err := os.OpenFile(
		filepath.Join(logDir, "error.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0666,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log file: %v", err)
	}

	return &Logger{
		commonLog: log.New(commonLogFile, "", 0),
		errorLog:  log.New(errorLogFile, "", 0),
	}, nil
}

// This function returns the package and function name where the log was called
func getPackageInfo() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	return fn.Name()
}

// This function formats the log message with timestamp, level, and package info
func (l *Logger) formatLog(level LogLevel, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, message)
}

func (l *Logger) Info(message string) {
	formattedMessage := l.formatLog(INFO, message)
	l.commonLog.Println(formattedMessage)
}

func (l *Logger) Warn(message string) {
	formattedMessage := l.formatLog(WARN, message)
	l.commonLog.Println(formattedMessage)
}

func (l *Logger) Error(message string) {
	formattedMessage := l.formatLog(ERROR, message)
	l.commonLog.Println(formattedMessage)
	l.errorLog.Println(formattedMessage)
}
