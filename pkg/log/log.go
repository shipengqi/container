package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/shipengqi/container/pkg/utils"
)

type Config struct {
	Level  string
	Output string
	Prefix string
}

const (
	DefaultLogDir    = "/var/run/log/container"
	DefaultLogLevel  = "debug"
	DefaultLogPrefix = "container"
)

var logger zerolog.Logger
var Output string

func Init(c *Config) (string, error) {
	if c.Output == "" {
		c.Output = DefaultLogDir
	}
	if c.Level == "" {
		c.Level = DefaultLogLevel
	}
	if c.Prefix == "" {
		c.Prefix = DefaultLogPrefix
	}

	logLevel := convertLogLevel(c.Level)
	zerolog.SetGlobalLevel(logLevel)

	// log output to files as well
	var w io.Writer

	logfile, err := normalizeFileName(c.Output, c.Prefix)
	if err != nil {
		return "", err
	}
	f, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: utils.IsWindows()}
	} else {
		w = zerolog.MultiLevelWriter(
			zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339, NoColor: utils.IsWindows()},
			f,
		)
	}

	logger = zerolog.New(w).With().Timestamp().Logger()
	Output = logfile
	return logfile, nil
}

func normalizeFileName(out, prefix string) (string, error) {
	if !utils.IsDir(out) {
		if err := os.MkdirAll(out, 0644); err != nil {
			return "", err
		}
	}
	logFileName := filepath.Join(out, fmt.Sprintf("%s.%s.log", prefix, time.Now().Format("20060102150405")))
	return logFileName, nil
}

func convertLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.DebugLevel
	}
}

func Trace() *zerolog.Event {
	return logger.Trace()
}

func Debug() *zerolog.Event {
	return logger.Debug()
}

func Info() *zerolog.Event {
	return logger.Info()
}

func Warn() *zerolog.Event {
	return logger.Warn()
}

func Error() *zerolog.Event {
	return logger.Error()
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func Panic() *zerolog.Event {
	return logger.Panic()
}
