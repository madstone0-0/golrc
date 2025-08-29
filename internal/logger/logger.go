package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var DEBUG bool = false

type Level int

const (
	Debug Level = iota
	Info
	Warning
	Error
	Fatal
)

var levelColors = map[slog.Level]string{
	slog.LevelDebug: "\x1b[37m", // white
	slog.LevelInfo:  "\x1b[32m", // green
	slog.LevelWarn:  "\x1b[33m", // yellow
	slog.LevelError: "\x1b[31m", // red
}

const reset = "\x1b[0m"

type Logger interface {
	I(format string, args ...any)
	W(format string, args ...any)
	E(format string, args ...any)
	D(format string, args ...any)
	F(format string, args ...any)
}

type TaggedLoggerHandler struct {
	Tag   string
	Attrs []slog.Attr
}

func (tlh TaggedLoggerHandler) Handle(ctx context.Context, r slog.Record) error {
	col, ok := levelColors[r.Level]
	if !ok {
		col = ""
	}

	// build prefix in color
	prefix := fmt.Sprintf("%s[%c]%s", col, r.Level.String()[0], reset)
	if tlh.Tag != "" {
		prefix = fmt.Sprintf("%s%s%s %s", col, tlh.Tag, reset, prefix)
	}

	// Gather all attrs: those on the handler + those on the record
	all := make([]slog.Attr, 0, len(tlh.Attrs)+r.NumAttrs())
	all = append(all, tlh.Attrs...)
	r.Attrs(func(a slog.Attr) bool {
		all = append(all, a)
		return true
	})

	// Format attrs as "key=value"
	var kvs []string
	for _, a := range all {
		kvs = append(kvs, fmt.Sprintf("%s=%v", a.Key, a.Value))
	}

	// Compose final message
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(" ")
	b.WriteString(r.Message)
	if len(kvs) > 0 {
		b.WriteString(" | ")
		b.WriteString(strings.Join(kvs, " "))
	}

	_, err := os.Stdout.WriteString(b.String() + "\n")
	return err
}

func (tlh TaggedLoggerHandler) Enabled(context.Context, slog.Level) bool {
	return true
}

func (tlh TaggedLoggerHandler) WithGroup(name string) slog.Handler {
	newTag := name
	if tlh.Tag != "" {
		newTag = tlh.Tag + "." + name
	}
	return &TaggedLoggerHandler{
		Tag:   newTag,
		Attrs: tlh.Attrs,
	}
}

func (tlh TaggedLoggerHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(tlh.Attrs), len(tlh.Attrs)+len(attrs))
	copy(newAttrs, tlh.Attrs)
	newAttrs = append(newAttrs, attrs...)
	return &TaggedLoggerHandler{
		Tag:   tlh.Tag,
		Attrs: newAttrs,
	}
}

type TaggedLogger struct {
	Handler TaggedLoggerHandler
	Logger  *slog.Logger
}

func (tl TaggedLogger) L(level Level, format string, args ...any) {
	switch level {
	case Debug:
		if DEBUG {
			tl.Logger.Debug(format, args...)
		}
	case Error:
		tl.Logger.Error(format, args...)
	case Fatal:
		tl.Logger.Error(format, args...)
		os.Exit(1)
	case Info:
		tl.Logger.Info(format, args...)
	case Warning:
		tl.Logger.Warn(format, args...)
	default:
		panic(fmt.Sprintf("unexpected logger.Level: %#v", level))
	}
}

func (tl TaggedLogger) I(format string, args ...any) {
	tl.L(Info, format, args...)
}

func (tl TaggedLogger) W(format string, args ...any) {
	tl.L(Warning, format, args...)
}

func (tl TaggedLogger) E(format string, args ...any) {
	tl.L(Error, format, args...)
}

func (tl TaggedLogger) D(format string, args ...any) {
	tl.L(Debug, format, args...)
}

func (tl TaggedLogger) F(format string, args ...any) {
	tl.L(Fatal, format, args...)
}

func NewTaggedLogger(tag string) TaggedLogger {
	handler := TaggedLoggerHandler{
		Tag: tag,
	}
	logger := slog.New(handler)
	return TaggedLogger{
		Handler: handler,
		Logger:  logger,
	}
}

var DefaultLogger Logger = NewTaggedLogger("")

func I(format string, args ...any) {
	DefaultLogger.I(format, args...)
}

func W(format string, args ...any) {
	DefaultLogger.W(format, args...)
}

func E(format string, args ...any) {
	DefaultLogger.E(format, args...)
}

func D(format string, args ...any) {
	if DEBUG {
		DefaultLogger.D(format, args...)
	}
}

func F(format string, args ...any) {
	DefaultLogger.F(format, args...)
}
