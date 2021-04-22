package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type ctxKey int

const (
	_ ctxKey = iota
	ctxLogCont
)

type ctxValues struct {
	values map[string]interface{}
}

func ContextWithValue(ctx context.Context, key string, value interface{}) context.Context {
	result := copyContextData(ctx)

	result.values[key] = value

	return context.WithValue(ctx, ctxLogCont, &result)
}

func ContextWithError(ctx context.Context, err error) context.Context {
	result := copyContextData(ctx)

	result.values["error"] = err.Error()

	return context.WithValue(ctx, ctxLogCont, &result)
}

func copyContextData(ctx context.Context) *ctxValues {
	result := ctxValues{values: map[string]interface{}{}}

	vals, _ := ctx.Value(ctxLogCont).(*ctxValues)
	if vals != nil {
		for k, v := range vals.values {
			result.values[k] = v
		}
	}

	return &result
}

type severity string

const (
	infoSeverity  severity = "INFO"
	errorSeverity severity = "ERROR"
)

type logOutput struct {
	Level   severity               `json:"level"`
	Context map[string]interface{} `json:"context,omitempty"`
	Message string                 `json:"message"`
}

type Logger interface {
	Info(ctx context.Context, template string, params ...interface{})
	Error(ctx context.Context, template string, params ...interface{})
}

func NewLogger() Logger {
	return logger{}
}

type logger struct{}

func (l logger) Info(ctx context.Context, template string, params ...interface{}) {
	l.log(ctx, infoSeverity, template, params...)
}

func (l logger) Error(ctx context.Context, template string, params ...interface{}) {
	l.log(ctx, errorSeverity, template, params...)
}

func (l logger) log(ctx context.Context, level severity, template string, params ...interface{}) {
	output := logOutput{
		Level:   level,
		Message: fmt.Sprintf(template, params...),
	}

	vals, _ := ctx.Value(ctxLogCont).(*ctxValues)
	if vals != nil {
		output.Context = vals.values
	}

	bytes, err := json.Marshal(&output)
	if err != nil {
		fmt.Printf("error logging: %s\n", err.Error())
	} else {
		fmt.Printf("%s -> %s", time.Now().Format(time.RFC3339), string(bytes))
	}
}
