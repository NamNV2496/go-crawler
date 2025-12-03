package logging

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
)

type Logger struct {
	*log.Logger
}

var loggerInstance *Logger

func init() {
	loggerInstance = NewLogger(log.Default())
}

func NewLogger(logger *log.Logger) *Logger {
	return &Logger{logger}
}

func GetLogger() *Logger {
	return loggerInstance
}

func (l *Logger) ResetPrefix(funcName string) {
	l.SetPrefix("[" + funcName + "] ")
}

func (l *Logger) AppendPrefix(funcName string) {
	l.SetPrefix(l.Prefix() + "[" + funcName + "] ")
}

func (l *Logger) Info(ctx context.Context, msg string) {
	traceID := ctx.Value("trace_id")
	l.Printf("[INFO] trace_id=%v - %s\n", traceID, msg)
}

func (l *Logger) Error(ctx context.Context, msg string) {
	traceID := ctx.Value("trace_id")
	l.Printf("[ERROR] trace_id=%v - %s\n", traceID, msg)
}

func (l *Logger) Debug(ctx context.Context, msg string) {
	traceID := ctx.Value("trace_id")
	l.Printf("[DEBUG] trace_id=%v - %s\n", traceID, msg)
}

func (l *Logger) Warn(ctx context.Context, msg string) {
	traceID := ctx.Value("trace_id")
	l.Printf("[WARN] trace_id=%v - %s\n", traceID, msg)
}

func Info(ctx context.Context, template string, v ...any) {
	msg := fmt.Sprintf(template, v...)
	loggerInstance.Info(ctx, msg)
}

func Error(ctx context.Context, template string, v ...any) {
	msg := fmt.Sprintf(template, v...)
	loggerInstance.Error(ctx, msg)
}

func Debug(ctx context.Context, template string, v ...any) {
	msg := fmt.Sprintf(template, v...)
	loggerInstance.Debug(ctx, msg)
}

func Warn(ctx context.Context, template string, v ...any) {
	msg := fmt.Sprintf(template, v...)
	loggerInstance.Warn(ctx, msg)
}

func AppendPrefix(funcName string) func() {
	previous := loggerInstance.Prefix()
	loggerInstance.AppendPrefix(funcName)
	return func() {
		loggerInstance.SetPrefix(previous)
	}
}

func ResetPrefix(ctx context.Context, funcName string) {
	loggerInstance.ResetPrefix(funcName)
}

func InjectTraceId(ctx context.Context) context.Context {
	return context.WithValue(ctx, "trace_id", uuid.New())
}
