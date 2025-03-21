package lokilogger

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFieldsToMap(t *testing.T) {
	err := fmt.Errorf("test error")
	stringer := &testStringer{value: "test stringer"}

	tests := []struct {
		name     string
		fields   []zap.Field
		expected map[string]string
	}{
		{
			name: "string field",
			fields: []zap.Field{
				zap.String("key", "value"),
			},
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name: "int field",
			fields: []zap.Field{
				zap.Int64("key", 42),
			},
			expected: map[string]string{
				"key": "42",
			},
		},
		{
			name: "float field",
			fields: []zap.Field{
				zap.Float64("key", 42.5),
			},
			expected: map[string]string{
				"key": "42.5",
			},
		},
		{
			name: "bool field",
			fields: []zap.Field{
				zap.Bool("key", true),
			},
			expected: map[string]string{
				"key": "true",
			},
		},
		{
			name: "duration field",
			fields: []zap.Field{
				zap.Duration("key", time.Second*5),
			},
			expected: map[string]string{
				"key": "5s",
			},
		},
		{
			name: "error field",
			fields: []zap.Field{
				zap.Error(err),
			},
			expected: map[string]string{
				"error": "test error",
			},
		},
		{
			name: "stringer field",
			fields: []zap.Field{
				zap.Stringer("key", stringer),
			},
			expected: map[string]string{
				"key": "test stringer",
			},
		},
		{
			name: "reflect field",
			fields: []zap.Field{
				zap.Any("key", map[string]int{"a": 1}),
			},
			expected: map[string]string{
				"key": "map[a:1]",
			},
		},
		{
			name: "multiple fields",
			fields: []zap.Field{
				zap.String("str", "value"),
				zap.Int64("int", 42),
				zap.Bool("bool", true),
			},
			expected: map[string]string{
				"str":  "value",
				"int":  "42",
				"bool": "true",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldsToMap(tt.fields)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("fieldsToMap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

type testStringer struct {
	value string
}

func (s *testStringer) String() string {
	return s.value
}
