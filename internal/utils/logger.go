package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

type Logger struct {
	level   LogLevel
	service string
}

type Field struct {
	Key   string
	Value interface{}
}

func NewLogger(service string) *Logger {
	level := getLogLevel()
	return &Logger{
		level:   level,
		service: service,
	}
}

func getLogLevel() LogLevel {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level == "" {
		level = "INFO"
	}
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return level >= l.level
}

func (l *Logger) formatMessage(level LogLevel, msg string, fields ...Field) string {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	formatted := fmt.Sprintf("[%s] %s [%s] %s", timestamp, level.String(), l.service, msg)

	if len(fields) > 0 {
		var fieldStrs []string
		for _, field := range fields {
			fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", field.Key, field.Value))
		}
		formatted += " " + strings.Join(fieldStrs, " ")
	}

	return formatted
}

func (l *Logger) Debug(msg string, fields ...Field) {
	if l.shouldLog(DEBUG) {
		log.Println(l.formatMessage(DEBUG, msg, fields...))
	}
}

func (l *Logger) Info(msg string, fields ...Field) {
	if l.shouldLog(INFO) {
		log.Println(l.formatMessage(INFO, msg, fields...))
	}
}

func (l *Logger) Warn(msg string, fields ...Field) {
	if l.shouldLog(WARN) {
		log.Println(l.formatMessage(WARN, msg, fields...))
	}
}

func (l *Logger) Error(msg string, fields ...Field) {
	if l.shouldLog(ERROR) {
		log.Println(l.formatMessage(ERROR, msg, fields...))
	}
}

func (l *Logger) ErrorWithErr(msg string, err error, fields ...Field) {
	if l.shouldLog(ERROR) {
		allFields := append(fields, Field{Key: "error", Value: err.Error()})
		log.Println(l.formatMessage(ERROR, msg, allFields...))
	}
}

var globalLogger *Logger

func InitGlobalLogger(service string) {
	globalLogger = NewLogger(service)
}

func GetLogger(service string) *Logger {
	if globalLogger == nil {
		InitGlobalLogger(service)
	}
	return globalLogger
}

func LogInfo(msg string) {
	logger := GetLogger("utils")
	logger.Info(msg)
}

func LogDebug(msg string) {
	logger := GetLogger("utils")
	logger.Debug(msg)
}