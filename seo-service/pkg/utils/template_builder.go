package utils

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
)

func BuildByTemplate(ctx context.Context, name, tpl string, request map[string]string) (string, error) {
	t, err := template.New(name).Option("missingkey=zero").Parse(tpl)
	if err != nil {
		return "", err
	}
	result := new(bytes.Buffer)
	err = t.Execute(result, request)
	if err != nil {
		return "", err
	}
	slog.Info("BuildByTemplate: ", "result", result.String())
	return result.String(), nil
}
