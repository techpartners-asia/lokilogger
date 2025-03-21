package lokilogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	service    string
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
		service: config.Service,
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

// fieldsToMap converts Zap fields to a map[string]string
func fieldsToMap(fields []zap.Field) map[string]string {
	result := make(map[string]string)
	for _, field := range fields {
		switch field.Type {
		case zapcore.StringType:
			result[field.Key] = field.String
		case zapcore.Int64Type:
			result[field.Key] = fmt.Sprintf("%d", field.Integer)
		case zapcore.Int32Type:
			result[field.Key] = fmt.Sprintf("%d", int32(field.Integer))
		case zapcore.Int16Type:
			result[field.Key] = fmt.Sprintf("%d", int16(field.Integer))
		case zapcore.Int8Type:
			result[field.Key] = fmt.Sprintf("%d", int8(field.Integer))
		case zapcore.Uint64Type:
			result[field.Key] = fmt.Sprintf("%d", uint64(field.Integer))
		case zapcore.Uint32Type:
			result[field.Key] = fmt.Sprintf("%d", uint32(field.Integer))
		case zapcore.Uint16Type:
			result[field.Key] = fmt.Sprintf("%d", uint16(field.Integer))
		case zapcore.Uint8Type:
			result[field.Key] = fmt.Sprintf("%d", uint8(field.Integer))
		case zapcore.Float64Type:
			result[field.Key] = fmt.Sprintf("%.1f", math.Float64frombits(uint64(field.Integer)))
		case zapcore.Float32Type:
			result[field.Key] = fmt.Sprintf("%.1f", math.Float32frombits(uint32(field.Integer)))
		case zapcore.BoolType:
			result[field.Key] = fmt.Sprintf("%v", field.Integer == 1)
		case zapcore.DurationType:
			result[field.Key] = time.Duration(field.Integer).String()
		case zapcore.TimeType:
			if field.Interface != nil {
				result[field.Key] = time.Unix(0, field.Integer).In(field.Interface.(*time.Location)).String()
			} else {
				result[field.Key] = time.Unix(0, field.Integer).String()
			}
		case zapcore.TimeFullType:
			result[field.Key] = field.Interface.(time.Time).String()
		case zapcore.ErrorType:
			result[field.Key] = field.Interface.(error).Error()
		case zapcore.StringerType:
			result[field.Key] = field.Interface.(fmt.Stringer).String()
		case zapcore.ReflectType:
			result[field.Key] = fmt.Sprintf("%v", field.Interface)
		case zapcore.ArrayMarshalerType:
			result[field.Key] = fmt.Sprintf("%v", field.Interface)
		case zapcore.ObjectMarshalerType:
			result[field.Key] = fmt.Sprintf("%v", field.Interface)
		case zapcore.InlineMarshalerType:
			result[field.Key] = fmt.Sprintf("%v", field.Interface)
		case zapcore.BinaryType:
			result[field.Key] = fmt.Sprintf("%x", field.Interface.([]byte))
		case zapcore.ByteStringType:
			result[field.Key] = fmt.Sprintf("%x", field.Interface.([]byte))
		case zapcore.Complex128Type:
			result[field.Key] = fmt.Sprintf("%v", field.Interface.(complex128))
		case zapcore.Complex64Type:
			result[field.Key] = fmt.Sprintf("%v", field.Interface.(complex64))
		case zapcore.UintptrType:
			result[field.Key] = fmt.Sprintf("%d", uintptr(field.Integer))
		case zapcore.NamespaceType:
			// Skip namespace fields as they don't have a direct string representation
		case zapcore.SkipType:
			// Skip skip fields
		}
	}
	return result
}

// sendLog sends a log entry to Loki
func (l *Logger) sendLog(entry LogEntry) error {
	// Create structured log entry
	logger := l.logger.With(entry.Fields...)

	// Log locally using Zap
	logger.Info(entry.Message)

	// Prepare Loki payload

	// Format the complete log message with all fields
	var messageBuilder bytes.Buffer
	messageBuilder.WriteString(entry.Message)
	if len(entry.Fields) > 0 {
		messageBuilder.WriteString(" | ")
		for i, field := range entry.Fields {
			if i > 0 {
				messageBuilder.WriteString(" ")
			}
			messageBuilder.WriteString(field.Key)
			messageBuilder.WriteString("=")
			messageBuilder.WriteString(fieldsToMap([]zap.Field{field})[field.Key])
		}
	}

	payload := LokiPayload{
		Streams: []Stream{
			{
				Stream: map[string]string{
					"source": l.service,
				},
				Values: [][]string{
					{
						fmt.Sprintf("%d", entry.Timestamp.UnixNano()),
						messageBuilder.String(),
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
