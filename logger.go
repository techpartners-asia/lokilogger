package lokilogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// LogEntry represents a single log entry with timestamp and message
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message"`
	Fields    []zap.Field
}

// LokiPayload represents the structure expected by Loki's HTTP API
type LokiPayload struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Logger handles communication with Loki and local logging
type Logger struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// Config holds the configuration for the logger
type Config struct {
	BaseURL     string
	Environment string
	Service     string
}

// New creates a new Logger instance
func New(config Config) (*Logger, error) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.Encoding = "json"

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{
		baseURL: config.BaseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.With(
			zap.String("service", config.Service),
			zap.String("environment", config.Environment),
		),
	}, nil
}

// Info logs an info message and sends it to Loki
func (l *Logger) Info(msg string, fields ...zap.Field) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}

	return l.sendLog(entry)
}

// Error logs an error message and sends it to Loki
func (l *Logger) Error(msg string, err error, fields ...zap.Field) error {
	fields = append(fields, zap.Error(err))
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}

	return l.sendLog(entry)
}

// Debug logs a debug message and sends it to Loki
func (l *Logger) Debug(msg string, fields ...zap.Field) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}

	return l.sendLog(entry)
}

// Warn logs a warning message and sends it to Loki
func (l *Logger) Warn(msg string, fields ...zap.Field) error {
	entry := LogEntry{
		Timestamp: time.Now(),
		Message:   msg,
		Fields:    fields,
	}

	return l.sendLog(entry)
}

// sendLog sends a log entry to Loki
func (l *Logger) sendLog(entry LogEntry) error {
	// Create structured log entry
	logger := l.logger.With(entry.Fields...)

	// Log locally using Zap
	logger.Info(entry.Message)

	// Prepare Loki payload
	payload := LokiPayload{
		Streams: []Stream{
			{
				Stream: map[string]string{
					"source": "lokilogger",
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
						entry.Message,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal payload", zap.Error(err))
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", l.baseURL+"/loki/api/v1/push", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create request", zap.Error(err))
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to send request", zap.Error(err))
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		logger.Error("Unexpected status code",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status))
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
