package logger

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/contrib/bridges/otelzap"
	"go.opentelemetry.io/otel/log/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type SentryConfig struct {
	Enable            bool
	DSN               string
	Environment       string
	ServerName        string
	Debug             bool // 是否开启调试模式
	Level             string
	Tags              map[string]string
	DisableStacktrace bool
	FlushTimeout      time.Duration
}

// SentryCoreConfig 定义 Sentry Core 的配置参数
type SentryCoreConfig struct {
	Tags              map[string]string // 用于 Sentry 上的标签
	DisableStacktrace bool              // 是否禁用堆栈跟踪
	Level             zapcore.Level     // 采集的日志级别
	FlushTimeout      time.Duration     // 刷新超时时间
	Hub               *sentry.Hub       // Sentry Hub 实例
}
type sentryCore struct {
	client               *sentry.Client         // sentry 客户端
	cfg                  *SentryCoreConfig      // 配置
	zapcore.LevelEnabler                        // 日志级别启用器
	flushTimeout         time.Duration          // 刷新超时
	fields               map[string]interface{} // 用于存储日志字段
}

// InitLoggerWithSentryConfig
type InitLoggerConfig struct {
	LogPath         string
	LogLevel        string
	SentryConfig    SentryConfig
	EnableOtelTrace bool
	EnableOtelLog   bool
}

// InitLoggerWithSentry 初始化带有 Sentry 的日志
// logPath 日志目录
// logLevel 日志级别
// sentryConfig Sentry 配置
// enableTrace 是否开启 trace
func InitLoggerWithSentry(cfg InitLoggerConfig) (*Logger, error) {
	// 如果未启用 Sentry，则直接初始化控制台日志
	if !cfg.SentryConfig.Enable {
		return InitConsoleLogger(cfg.LogLevel, cfg.EnableOtelTrace)
	}
	// 初始化 Sentry 客户端
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:   cfg.SentryConfig.DSN,
		Debug: cfg.SentryConfig.Debug,
	}); err != nil {
		return nil, fmt.Errorf("sentry initialization failed: %w", err)
	}

	// 创建 SentryCore 配置
	sentryCoreConfig := SentryCoreConfig{
		Level: getLevel(cfg.SentryConfig.Level),
		Tags: map[string]string{
			"environment": cfg.SentryConfig.Environment,
			"server_name": cfg.SentryConfig.ServerName,
		},
	}

	// 创建 SentryCore
	sentryClient := sentry.CurrentHub().Client()
	sentryCore := NewSentryCore(sentryCoreConfig, sentryClient)

	// 初始化普通日志
	logger, err := InitConsoleLogger(cfg.LogLevel, cfg.EnableOtelTrace)
	if err != nil {
		return nil, err
	}

	// 将 SentryCore 添加到现有 Logger
	logger.l = logger.l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.NewTee(core, sentryCore)
	}))
	if cfg.EnableOtelLog {
		logPusher := global.GetLoggerProvider()
		// 将 logPusher 添加到 logger
		logger.l = logger.l.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, otelzap.NewCore(cfg.SentryConfig.ServerName, otelzap.WithLoggerProvider(logPusher)))
		}))
	}

	defer logger.Sync()
	return logger, nil
}

// NewSentryCore 创建 Sentry 核心
func NewSentryCore(cfg SentryCoreConfig, sentryClient *sentry.Client) zapcore.Core {
	return &sentryCore{
		client:       sentryClient,
		cfg:          &cfg,
		LevelEnabler: cfg.Level,
		flushTimeout: 3 * time.Second, // 默认值，超时3秒
		fields:       make(map[string]interface{}),
	}
}

// With 方法用于合并额外的字段
func (c *sentryCore) With(fs []zapcore.Field) zapcore.Core {
	m := make(map[string]interface{}, len(c.fields))
	for k, v := range c.fields {
		m[k] = v
	}

	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fs {
		f.AddTo(enc)
	}

	for k, v := range enc.Fields {
		m[k] = v
	}

	return &sentryCore{
		client:       c.client,
		cfg:          c.cfg,
		fields:       m,
		LevelEnabler: c.LevelEnabler,
	}
}

// Write 实现 zapcore.Core 接口的 Write 方法
// 这个方法会把日志写到 Sentry
func (c *sentryCore) Write(ent zapcore.Entry, fs []zapcore.Field) error {
	clone := c.With(fs)

	event := sentry.NewEvent()
	event.Message = ent.Message
	event.Timestamp = ent.Time
	event.Level = sentryLevel(ent.Level) // 将 zap 的 Level 转换为 sentry 的 Level
	event.Platform = "zap"
	event.Extra = clone.(*sentryCore).fields
	event.Tags = c.cfg.Tags

	// 是否禁用堆栈跟踪
	if !c.cfg.DisableStacktrace {
		trace := sentry.NewStacktrace()
		if trace != nil {
			event.Exception = []sentry.Exception{{
				Type:       ent.Message,
				Value:      ent.Caller.TrimmedPath(),
				Stacktrace: trace,
			}}
		}
	}

	hub := c.cfg.Hub
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	_ = c.client.CaptureEvent(event, nil, hub.Scope())

	// 如果是 Error 以上的级别，刷新 Sentry 客户端
	if ent.Level >= zapcore.ErrorLevel {
		c.client.Flush(c.flushTimeout)
	}
	return nil
}

// Check 实现 zapcore.Core 接口的 Check 方法
func (c *sentryCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.cfg.Level.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Sync 实现 zapcore.Core 接口的 Sync 方法
func (c *sentryCore) Sync() error {
	c.client.Flush(c.flushTimeout)
	return nil
}

// sentryLevel 将 zapcore.Level 转换为 sentry.Level
func sentryLevel(lvl zapcore.Level) sentry.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return sentry.LevelDebug
	case zapcore.InfoLevel:
		return sentry.LevelInfo
	case zapcore.WarnLevel:
		return sentry.LevelWarning
	case zapcore.ErrorLevel:
		return sentry.LevelError
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelFatal
	}
}

// getLevel
func getLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}
