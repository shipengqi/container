package log

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	configureMutex sync.Mutex
	// loggingConfigured will be set once logging has been configured via invoking `ConfigureLogging`.
	// Subsequent invocations of `ConfigureLogging` would be no-op
	loggingConfigured = false
)

var Output string

type Config struct {
	FileLevel string
	// Directory to log to dir when filelogging is enabled
	Directory string
	// Filename is the name of the logfile which will be placed inside the directory
	Filename string
}

type Logger struct {
	Unsugared *zap.Logger
	*zap.SugaredLogger
}

// Debugt Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Debugt(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Debug(msg, fields...)
}

func (logger *Logger) Debug(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Debugw(msg, keyValues...)
}

func (logger *Logger) Debugs(args ...interface{}) {
	logger.SugaredLogger.Debug(args...)
}

// Infot Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Infot(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Info(msg, fields...)
}

func (logger *Logger) Info(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Infow(msg, keyValues...)
}

func (logger *Logger) Infos(args ...interface{}) {
	logger.SugaredLogger.Info(args...)
}

// Warnt Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Warnt(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Warn(msg, fields...)
}

func (logger *Logger) Warn(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Warnw(msg, keyValues...)
}

func (logger *Logger) Warns(args ...interface{}) {
	logger.SugaredLogger.Warn(args...)
}

// Errort Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Errort(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Error(msg, fields...)
}

func (logger *Logger) Error(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Errorw(msg, keyValues...)
}

func (logger *Logger) Errors(args ...interface{}) {
	logger.SugaredLogger.Error(args...)
}

// Panict Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Panict(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Panic(msg, fields...)
}

func (logger *Logger) Panic(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Panicw(msg, keyValues...)
}

func (logger *Logger) Panics(args ...interface{}) {
	logger.SugaredLogger.Panic(args...)
}

// Fatalt Use zap.String(key, value), zap.Int(key, value) to log fields. These fields
// will be marshalled as JSON in the logfile and key value pairs in the console!
func (logger *Logger) Fatalt(msg string, fields ...zapcore.Field) {
	logger.Unsugared.Fatal(msg, fields...)
}

func (logger *Logger) Fatal(msg string, keyValues ...interface{}) {
	logger.SugaredLogger.Fatalw(msg, keyValues...)
}

func (logger *Logger) Fatals(args ...interface{}) {
	logger.SugaredLogger.Fatal(args...)
}

// Examples:
// logger.Infot("Importing new file", zap.String("source", filename), zap.Int("size", 1024))
// logger.Info("Importing new file", "source", filename, "size", 1024)
// To log a stacktrace:
// logger.Errort("It went wrong, zap.Stack())

// defaultZapLogger is the default logger instance that should be used to log
// It's assigned a default value here for tests (which do not call log.Configure())
var defaultZapLogger *Logger

func init() {
	defaultZapLogger, _ = newZapLogger(Config{
		FileLevel: "DEBUG",
	})
}

func Debugt(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Debugt(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	defaultZapLogger.Debugf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Debugw(msg, keysAndValues...)
}

func Debug(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Debug(msg, keysAndValues...)
}

func Debugs(args ...interface{}) {
	defaultZapLogger.Debugs(args...)
}

func Infot(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Infot(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	defaultZapLogger.Infof(template, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Infow(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Info(msg, keysAndValues...)
}

func Infos(args ...interface{}) {
	defaultZapLogger.Infos(args...)
}

func Warnt(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Warnt(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	defaultZapLogger.Warnf(template, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Warnw(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Warn(msg, keysAndValues...)
}

func Warns(args ...interface{}) {
	defaultZapLogger.Warns(args...)
}

func Errort(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Errort(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	defaultZapLogger.Errorf(template, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Errorw(msg, keysAndValues...)
}

func Error(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Error(msg, keysAndValues...)
}

func Errors(args ...interface{}) {
	defaultZapLogger.Errors(args...)
}

func Panict(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Panict(msg, fields...)
}

func Panicf(template string, args ...interface{}) {
	defaultZapLogger.Panicf(template, args...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Panicw(msg, keysAndValues...)
}

func Panic(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Panic(msg, keysAndValues...)
}

func Panics(args ...interface{}) {
	defaultZapLogger.Panics(args...)
}

func Fatalt(msg string, fields ...zapcore.Field) {
	defaultZapLogger.Fatalt(msg, fields...)
}

func Fatalf(template string, args ...interface{}) {
	defaultZapLogger.Fatalf(template, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Fatalw(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	defaultZapLogger.Fatal(msg, keysAndValues...)
}

func Fatals(args ...interface{}) {
	defaultZapLogger.Fatals(args...)
}

// Configure sets up the logging framework
//
// In production, the container logs will be collected and file logging should be disabled. However,
// during development it's nicer to see logs as text and optionally write to a file when debugging
// problems in the containerized pipeline
//
// The output log file will be located at /var/log/service-xyz/service-xyz.log and
// will be rolled according to configuration set.
func Configure(config Config) (*Logger, error) {
	configureMutex.Lock()
	defer configureMutex.Unlock()

	if loggingConfigured {
		return defaultZapLogger, nil
	}

	logger, err := newZapLogger(config)
	if err != nil {
		return nil, err
	}
	logger.Infot("logging configured",
		zap.String("fileLevel", config.FileLevel),
		zap.String("logDirectory", config.Directory),
		zap.String("fileName", config.Filename))

	defaultZapLogger = logger
	zap.RedirectStdLog(defaultZapLogger.Unsugared)
	loggingConfigured = true
	return logger, nil
}

func newZapLogger(config Config) (*Logger, error) {
	var fileLevel zapcore.Level
	err := fileLevel.Set(strings.ToLower(config.FileLevel))
	if err != nil {
		return nil, err
	}

	jsonEncCfg := zapcore.EncoderConfig{
		TimeKey:        "@timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
	}

	fileLevelEnabler := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= fileLevel
	})
	fileEncoder := zapcore.NewJSONEncoder(jsonEncCfg)

	if err := os.MkdirAll(config.Directory, 0744); err != nil {
		Error("can't create log directory", zap.Error(err), zap.String("path", config.Directory))
		return nil, err
	}

	file := filepath.Join(config.Directory, config.Filename)
	fd, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	core := zapcore.NewCore(fileEncoder, fd, fileLevelEnabler)
	unSugared := zap.New(core)
	return &Logger{
		Unsugared: unSugared,
		SugaredLogger: unSugared.Sugar(),
	}, nil
}
