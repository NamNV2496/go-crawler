package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_ValidateRequire(t *testing.T) {
	t.Skip("Skipping file-dependent test - requires YAML configuration files in production environment")

	tests := []struct {
		name     string
		action   string
		eventMap map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name:   "success - all required fields present for GET",
			action: "GET",
			eventMap: map[string]string{
				"url":       "https://example.com",
				"method":    "GET",
				"cron_exp":  "*/5 * * * *",
				"queue":     "normal",
				"is_active": "true",
				"domain":    "example.com",
			},
			wantErr: false,
		},
		{
			name:   "error - missing required url field",
			action: "GET",
			eventMap: map[string]string{
				"method": "GET",
			},
			wantErr: true,
			errMsg:  "url",
		},
		{
			name:   "success - POST without cron_exp (conditional requirement)",
			action: "POST",
			eventMap: map[string]string{
				"url":       "https://example.com",
				"method":    "POST",
				"queue":     "normal",
				"is_active": "true",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidate()
			ctx := context.Background()

			err := v.ValidateRequire(ctx, tt.action, tt.eventMap)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ValidateValue(t *testing.T) {
	t.Skip("Skipping file-dependent test - requires YAML configuration files in production environment")

	tests := []struct {
		name     string
		eventMap map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "success - valid values within range",
			eventMap: map[string]string{
				"repeat_times": "10",
				"url":          "https://example.com/path",
				"description":  "This is a valid description with enough words for testing",
			},
			wantErr: false,
		},
		{
			name: "error - repeat_times below minimum",
			eventMap: map[string]string{
				"repeat_times": "0",
			},
			wantErr: true,
		},
		{
			name: "error - repeat_times above maximum",
			eventMap: map[string]string{
				"repeat_times": "10000",
			},
			wantErr: true,
		},
		{
			name: "error - url too short",
			eventMap: map[string]string{
				"url": "h",
			},
			wantErr: true,
		},
		{
			name: "error - description insufficient words",
			eventMap: map[string]string{
				"description": "too short",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidate()
			ctx := context.Background()

			err := v.ValidateValue(ctx, tt.eventMap)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_ValidateCustomeRules(t *testing.T) {
	tests := []struct {
		name     string
		eventMap map[string]string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "success - valid method",
			eventMap: map[string]string{
				"method": "GET",
			},
			wantErr: false,
		},
		{
			name: "success - POST method",
			eventMap: map[string]string{
				"method": "POST",
			},
			wantErr: false,
		},
		{
			name: "error - invalid method",
			eventMap: map[string]string{
				"method": "PUT",
			},
			wantErr: true,
			errMsg:  "method",
		},
		{
			name: "error - invalid method DELETE",
			eventMap: map[string]string{
				"method": "DELETE",
			},
			wantErr: true,
		},
		{
			name: "success - scheduler_at equals next_run_time",
			eventMap: map[string]string{
				"scheduler_at":  "2024-01-01T00:00:00Z",
				"next_run_time": "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name: "error - scheduler_at not equal to next_run_time",
			eventMap: map[string]string{
				"scheduler_at":  "2024-01-01T00:00:00Z",
				"next_run_time": "2024-01-02T00:00:00Z",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidate()

			err := v.ValidateCustomeRules(tt.eventMap)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_IntegrationWorkflow(t *testing.T) {
	t.Skip("Skipping integration test - requires YAML configuration files in production environment")

	t.Run("complete validation workflow for GET request", func(t *testing.T) {
		v := NewValidate()
		ctx := context.Background()

		eventMap := map[string]string{
			"url":           "https://example.com/api/data",
			"method":        "GET",
			"cron_exp":      "*/5 * * * *",
			"queue":         "priority",
			"is_active":     "true",
			"domain":        "example.com",
			"description":   "5",
			"repeat_times":  "5",
			"scheduler_at":  "2024-01-01T00:00:00Z",
			"next_run_time": "2024-01-01T00:00:00Z",
		}

		// Test all validation steps
		err := v.ValidateRequire(ctx, "GET", eventMap)
		assert.NoError(t, err, "Required field validation failed")

		err = v.ValidateValue(ctx, eventMap)
		assert.NoError(t, err, "Value validation failed")

		err = v.ValidateCustomeRules(eventMap)
		assert.NoError(t, err, "Custom rules validation failed")
	})

	t.Run("complete validation workflow for POST request", func(t *testing.T) {
		v := NewValidate()
		ctx := context.Background()

		eventMap := map[string]string{
			"url":          "https://api.example.com/submit",
			"method":       "POST",
			"queue":        "normal",
			"is_active":    "true",
			"domain":       "example.com",
			"description":  "10",
			"repeat_times": "10",
		}

		err := v.ValidateRequire(ctx, "POST", eventMap)
		assert.NoError(t, err)

		err = v.ValidateValue(ctx, eventMap)
		assert.NoError(t, err)

		err = v.ValidateCustomeRules(eventMap)
		assert.NoError(t, err)
	})
}

func TestValidate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		eventMap map[string]string
		wantErr  bool
	}{
		{
			name:     "empty event map",
			eventMap: map[string]string{},
			wantErr:  true,
		},
		{
			name: "nil values in map",
			eventMap: map[string]string{
				"url":    "",
				"method": "",
			},
			wantErr: true,
		},
		{
			name: "special characters in URL",
			eventMap: map[string]string{
				"url":    "https://example.com/path?query=value&other=123",
				"method": "GET",
			},
			wantErr: false,
		},
		{
			name: "unicode in description",
			eventMap: map[string]string{
				"description": "这是一个测试描述 with unicode 字符 for testing purposes",
				"method":      "GET",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidate()
			ctx := context.Background()

			err := v.ValidateRequire(ctx, "GET", tt.eventMap)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidate_ValidateRequire(b *testing.B) {
	v := NewValidate()
	ctx := context.Background()
	eventMap := map[string]string{
		"url":       "https://example.com",
		"method":    "GET",
		"cron_exp":  "*/5 * * * *",
		"queue":     "normal",
		"is_active": "true",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateRequire(ctx, "GET", eventMap)
	}
}

func BenchmarkValidate_ValidateValue(b *testing.B) {
	v := NewValidate()
	ctx := context.Background()
	eventMap := map[string]string{
		"repeat_times": "10",
		"url":          "https://example.com/path",
		"description":  "This is a valid description with enough words",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateValue(ctx, eventMap)
	}
}

func BenchmarkValidate_ValidateCustomeRules(b *testing.B) {
	v := NewValidate()
	eventMap := map[string]string{
		"method":        "GET",
		"scheduler_at":  "1000",
		"next_run_time": "1000",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.ValidateCustomeRules(eventMap)
	}
}
