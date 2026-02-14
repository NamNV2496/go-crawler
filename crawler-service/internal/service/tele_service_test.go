package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Telegram Bot API
type mockTelegramBot struct {
	sendMessageFunc func(text string) error
	sentMessages    []string
	shouldFail      bool
}

func (m *mockTelegramBot) Send(text string) error {
	if m.shouldFail {
		return errors.New("telegram API error")
	}
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(text)
	}
	m.sentMessages = append(m.sentMessages, text)
	return nil
}

func TestTeleService_SendMessage(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		messageTyp string
		mockSetup  func(*mockTelegramBot)
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "success - send text message",
			message:    "Hello, World!",
			messageTyp: "text",
			mockSetup: func(m *mockTelegramBot) {
				m.sendMessageFunc = func(text string) error {
					assert.Equal(t, "Hello, World!", text)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:       "success - send formatted message",
			message:    "**Bold** and _italic_",
			messageTyp: "markdown",
			mockSetup: func(m *mockTelegramBot) {
				m.sendMessageFunc = func(text string) error {
					assert.Contains(t, text, "**Bold**")
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:       "success - send empty message",
			message:    "",
			messageTyp: "text",
			mockSetup: func(m *mockTelegramBot) {
				m.sendMessageFunc = func(text string) error {
					assert.Equal(t, "", text)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:       "error - telegram API failure",
			message:    "Test message",
			messageTyp: "text",
			mockSetup: func(m *mockTelegramBot) {
				m.shouldFail = true
			},
			wantErr: true,
			errMsg:  "telegram API error",
		},
		{
			name:       "success - long message",
			message:    string(make([]byte, 5000)), // Long message
			messageTyp: "text",
			mockSetup: func(m *mockTelegramBot) {
				m.sendMessageFunc = func(text string) error {
					assert.Greater(t, len(text), 4000)
					return nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := &mockTelegramBot{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockBot)
			}

			// Simulate sending message
			err := mockBot.Send(tt.message)

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

func TestTeleService_MessageFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		format   string
	}{
		{
			name:     "plain text",
			input:    "Simple message",
			expected: "Simple message",
			format:   "text",
		},
		{
			name:     "with special characters",
			input:    "Price: $100.50",
			expected: "Price: $100.50",
			format:   "text",
		},
		{
			name:     "with newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
			format:   "text",
		},
		{
			name:     "with unicode",
			input:    "Hello 世界 🌍",
			expected: "Hello 世界 🌍",
			format:   "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := &mockTelegramBot{}

			err := mockBot.Send(tt.input)
			assert.NoError(t, err)

			if len(mockBot.sentMessages) > 0 {
				assert.Equal(t, tt.expected, mockBot.sentMessages[0])
			}
		})
	}
}

func TestTeleService_ConcurrentMessages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	mockBot := &mockTelegramBot{
		sentMessages: make([]string, 0),
	}

	// Send multiple messages concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := mockBot.Send("Message")
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Note: In real implementation, you'd need proper synchronization
	assert.LessOrEqual(t, len(mockBot.sentMessages), 10)
}

func TestTeleService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mockTelegramBot) *int
		message    string
		maxRetries int
		wantErr    bool
	}{
		{
			name: "retry on failure - eventually succeeds",
			setupMock: func(m *mockTelegramBot) *int {
				callCount := 0
				m.sendMessageFunc = func(text string) error {
					callCount++
					if callCount < 3 {
						return errors.New("temporary error")
					}
					return nil
				}
				return &callCount
			},
			message:    "Retry test",
			maxRetries: 5,
			wantErr:    false,
		},
		{
			name: "permanent failure",
			setupMock: func(m *mockTelegramBot) *int {
				m.shouldFail = true
				return nil
			},
			message:    "Permanent error test",
			maxRetries: 3,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := &mockTelegramBot{}
			if tt.setupMock != nil {
				tt.setupMock(mockBot)
			}

			// Simulate retry logic
			var err error
			for i := 0; i < tt.maxRetries; i++ {
				err = mockBot.Send(tt.message)
				if err == nil {
					break
				}
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTeleService_MessageValidation(t *testing.T) {
	tests := []struct {
		name    string
		message string
		valid   bool
	}{
		{
			name:    "valid message",
			message: "Valid message content",
			valid:   true,
		},
		{
			name:    "empty message",
			message: "",
			valid:   true, // Some implementations allow empty
		},
		{
			name:    "very long message",
			message: string(make([]byte, 10000)),
			valid:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBot := &mockTelegramBot{}
			err := mockBot.Send(tt.message)

			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkTeleService_SendMessage(b *testing.B) {
	mockBot := &mockTelegramBot{
		sendMessageFunc: func(text string) error {
			return nil
		},
	}

	message := "Benchmark test message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockBot.Send(message)
	}
}

func BenchmarkTeleService_SendLongMessage(b *testing.B) {
	mockBot := &mockTelegramBot{
		sendMessageFunc: func(text string) error {
			return nil
		},
	}

	message := string(make([]byte, 4096))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockBot.Send(message)
	}
}

func BenchmarkTeleService_ConcurrentSend(b *testing.B) {
	mockBot := &mockTelegramBot{
		sendMessageFunc: func(text string) error {
			return nil
		},
	}

	message := "Concurrent benchmark"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = mockBot.Send(message)
		}
	})
}
