package logger

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

// 控制台日志
func InitConsoleLogger(logLevel string, enableTrace bool) (*Logger, error) {
	dynamicLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		hlog.Error("parse log level error", err)
		return nil, err
	}
	logger := NewLogger(
		WithTraceEnable(enableTrace),
		WithZapOptions(zap.AddCaller(),
			zap.AddCallerSkip(3)),
		WithCores([]CoreConfig{
			{
				Enc: zapcore.NewConsoleEncoder(humanEncoderConfig()),
				Ws:  zapcore.AddSync(os.Stdout),
				Lvl: dynamicLevel,
			},
		}...),
	)
	defer logger.Sync()
	return logger, nil
}

// InitLogger 胖日志初始化
// logPath 日志目录
// logLevel 日志级别
// 动态日志级别分类输出,开启行号显示
func InitLogger(logPath string, logLevel string, enableTrace bool) (*Logger, error) {
	// 创建日志目录
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	dynamicLevel, err := zap.ParseAtomicLevel(logLevel)
	if err != nil {
		hlog.Error("parse log level error", err)
		return nil, err
	}
	logger := NewLogger(
		WithTraceEnable(enableTrace),
		WithZapOptions(zap.AddCaller(),
			zap.AddCallerSkip(3)),
		WithCores([]CoreConfig{
			{
				Enc: zapcore.NewConsoleEncoder(humanEncoderConfig()),
				Ws:  zapcore.AddSync(os.Stdout),
				Lvl: dynamicLevel,
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer(logPath + "/all.log"),
				Lvl: zap.NewAtomicLevelAt(zapcore.DebugLevel),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer(logPath + "/debug.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.DebugLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer(logPath + "/info.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.InfoLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer(logPath + "/warn.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.WarnLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer(logPath + "/error.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev >= zap.ErrorLevel
				}),
			},
		}...),
	)
	defer logger.Sync()
	return logger, nil
}

// humanEncoderConfig copy from zap
func humanEncoderConfig() zapcore.EncoderConfig {
	cfg := commonEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	return cfg
}

func getWriteSyncer(file string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   file,
		MaxSize:    100,
		MaxBackups: 10,
		MaxAge:     7,
		Compress:   true,
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// testEncoderConfig encoder config for testing, copy from zap
func commonEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
		NameKey:    "name",
		TimeKey:    "time",
		CallerKey:  "file",
		//FunctionKey:    "func",
		StacktraceKey:  "stacktrace",
		LineEnding:     "\n",
		EncodeTime:     zapcore.EpochTimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}

func commonAccessEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{}
}
