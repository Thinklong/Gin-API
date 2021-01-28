package logger

import (
	"gin-api/internal/pkg/config"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// DefaultLevel the default log level
	DefaultLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
	// DefaultTimeLayout the default time layout;
	DefaultTimeLayout = time.RFC3339
)

// Option custom setup config
type Option func(*option)

type option struct {
	level      zapcore.Level
	fields     map[string]string
	file       io.Writer
	timeLayout string
}

// WithDebugLevel only greater than 'level' will output
func WithDebugLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.DebugLevel
	}
}

// WithInfoLevel only greater than 'level' will output
func WithInfoLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.InfoLevel
	}
}

// WithWarnLevel only greater than 'level' will output
func WithWarnLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.WarnLevel
	}
}

// WithErrorLevel only greater than 'level' will output
func WithErrorLevel() Option {
	return func(opt *option) {
		opt.level = zapcore.ErrorLevel
	}
}

// WithField add some field(s) to log
func WithField(key, value string) Option {
	return func(opt *option) {
		opt.fields[key] = value
	}
}

// WithFileP write log to some file
func WithFileP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0766)
	if err != nil {
		panic(err)
	}

	return func(opt *option) {
		opt.file = zapcore.Lock(f)
	}
}

// WithFileRotationP write log to some file with rotation
func WithFileRotationP(file string) Option {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0766); err != nil {
		panic(err)
	}
	logDir := config.GetConfig().GetString("log.fileDir")
	serviceName := config.GetConfig().GetString("project.name")
	moduleName := "main"

	return func(opt *option) {
		opt.file = &lumberjack.Logger{ // concurrent-safed
			Filename:   logDir + "/" + serviceName + "-" + moduleName + ".log", // 日志文件路径
			MaxSize:    1024,                                                   // 每个日志文件保存的最大尺寸 单位：M
			MaxBackups: 30,                                                     // 日志文件最多保存多少个备份
			MaxAge:     7,                                                      // 文件最多保存多少天
			Compress:   true,                                                   // 是否压缩
		}
	}
}

// WithTimeLayout custom time format
func WithTimeLayout(timeLayout string) Option {
	return func(opt *option) {
		opt.timeLayout = timeLayout
	}
}

// NewJSONLogger return a json-encoder zap logger,
func NewJSONLogger(opts ...Option) (*zap.Logger, error) {
	opt := &option{level: DefaultLevel, fields: make(map[string]string)}
	for _, f := range opts {
		f(opt)
	}

	// similar to zap.NewProductionEncoderConfig()
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "line",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, //
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// lowPriority usd by info\debug\warn
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl < zapcore.ErrorLevel
	})

	// highPriority usd by error\panic\fatal
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= opt.level && lvl >= zapcore.ErrorLevel
	})

	stdout := zapcore.Lock(os.Stdout) // lock for concurrent safe
	stderr := zapcore.Lock(os.Stderr) // lock for concurrent safe

	logDir := config.GetConfig().GetString("log.fileDir")
	serviceName := config.GetConfig().GetString("project.name")
	moduleName := "main"
	hook := lumberjack.Logger{
		Filename:   logDir + "/" + serviceName + "-" + moduleName + ".log", // 日志文件路径
		MaxSize:    1024,                                                   // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,                                                     // 日志文件最多保存多少个备份
		MaxAge:     7,                                                      // 文件最多保存多少天
		Compress:   true,                                                   // 是否压缩
	}

	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(&hook), opt.level),
		zapcore.NewCore(jsonEncoder,
			zapcore.NewMultiWriteSyncer(stdout),
			lowPriority,
		),
		zapcore.NewCore(jsonEncoder,
			zapcore.NewMultiWriteSyncer(stderr),
			highPriority,
		),
	)

	if opt.file != nil {
		core = zapcore.NewTee(core,
			zapcore.NewCore(jsonEncoder,
				zapcore.AddSync(opt.file),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= opt.level
				}),
			),
		)
	}

	logger := zap.New(core,
		zap.AddCaller(),
		zap.ErrorOutput(stderr),
	)

	for key, value := range opt.fields {
		logger = logger.WithOptions(zap.Fields(zapcore.Field{Key: key, Type: zapcore.StringType, String: value}))
	}
	return logger, nil
}

var _ Meta = (*meta)(nil)

// Meta key-value
type Meta interface {
	Key() string
	Value() interface{}
	meta()
}

type meta struct {
	key   string
	value interface{}
}

func (m *meta) Key() string {
	return m.key
}

func (m *meta) Value() interface{} {
	return m.value
}

func (m *meta) meta() {}

// NewMeta create meat
func NewMeta(key string, value interface{}) Meta {
	return &meta{key: key, value: value}
}

// WrapMeta wrap meta to zap fields
func WrapMeta(err error, metas ...Meta) (fields []zap.Field) {
	capacity := len(metas) + 1 // namespace meta
	if err != nil {
		capacity++
	}

	fields = make([]zap.Field, 0, capacity)
	if err != nil {
		fields = append(fields, zap.Error(err))
	}

	fields = append(fields, zap.Namespace("meta"))
	for _, meta := range metas {
		fields = append(fields, zap.Any(meta.Key(), meta.Value()))
	}

	return
}
