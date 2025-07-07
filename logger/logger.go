// Copyright 2022 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import (
	"context"
	"errors"
	"io"
	"runtime"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ hlog.FullLogger = (*Logger)(nil)

type Logger struct {
	l      *zap.Logger
	config *config
}

const (
	traceIDKey    = "trace_id"
	spanIDKey     = "span_id"
	traceFlagsKey = "trace_flags"
	// 添加代码位置相关的常量
	codeFilepathKey  = "code_filepath"
	codeFunctionKey  = "code_function"
	codeLinenoKey    = "code_lineno"
	codeNamespaceKey = "code_namespace"
)

var extraKeys = []ExtraKey{traceIDKey, spanIDKey, traceFlagsKey}

func NewLogger(opts ...Option) *Logger {
	cfg := defaultConfig()
	// apply options
	for _, opt := range opts {
		opt.apply(cfg)
	}
	cores := make([]zapcore.Core, 0, len(cfg.coreConfigs))
	for _, coreConfig := range cfg.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}
	loggerZap := zap.New(
		zapcore.NewTee(cores[:]...),
		cfg.zapOpts...)
	logger := Logger{
		l:      loggerZap,
		config: cfg,
	}
	logger.PutExtraKeys(extraKeys...)
	return &logger
}

// GetExtraKeys get extraKeys from logger config
func (l *Logger) GetExtraKeys() []ExtraKey {
	return l.config.extraKeys
}

// PutExtraKeys add extraKeys after init
func (l *Logger) PutExtraKeys(keys ...ExtraKey) {
	for _, k := range keys {
		if !InArray(k, l.config.extraKeys) {
			l.config.extraKeys = append(l.config.extraKeys, k)
		}
	}
}

func (l *Logger) Log(level hlog.Level, kvs ...interface{}) {
	sugar := l.l.Sugar()
	switch level {
	case hlog.LevelTrace, hlog.LevelDebug:
		sugar.Debug(kvs...)
	case hlog.LevelInfo:
		sugar.Info(kvs...)
	case hlog.LevelNotice, hlog.LevelWarn:
		sugar.Warn(kvs...)
	case hlog.LevelError:
		sugar.Error(kvs...)
	case hlog.LevelFatal:
		sugar.Fatal(kvs...)
	default:
		sugar.Warn(kvs...)
	}
}

func (l *Logger) Logf(level hlog.Level, format string, kvs ...interface{}) {
	logger := l.l.Sugar().With()
	switch level {
	case hlog.LevelTrace, hlog.LevelDebug:
		logger.Debugf(format, kvs...)
	case hlog.LevelInfo:
		logger.Infof(format, kvs...)
	case hlog.LevelNotice, hlog.LevelWarn:
		logger.Warnf(format, kvs...)
	case hlog.LevelError:
		logger.Errorf(format, kvs...)
	case hlog.LevelFatal:
		logger.Fatalf(format, kvs...)
	default:
		logger.Warnf(format, kvs...)
	}
}

func (l *Logger) ctxLogf(level hlog.Level, ctx context.Context, format string, kvs ...interface{}) {
	zLevel := hLevelToZapLevel(level)
	if !l.config.coreConfigs[0].Lvl.Enabled(zLevel) {
		return
	}
	zapLogger := l.l
	// 添加代码位置信息
	if filepath := ctx.Value(codeFilepathKey); filepath != nil {
		zapLogger = zapLogger.With(zap.String("code_filepath", filepath.(string)))
	}
	if funcName := ctx.Value(codeFunctionKey); funcName != nil {
		zapLogger = zapLogger.With(zap.String("code_function", funcName.(string)))
	}
	if lineno := ctx.Value(codeLinenoKey); lineno != nil {
		zapLogger = zapLogger.With(zap.Int("code_lineno", lineno.(int)))
	}
	if namespace := ctx.Value(codeNamespaceKey); namespace != nil {
		zapLogger = zapLogger.With(zap.String("code_namespace", namespace.(string)))
	}

	if len(l.config.extraKeys) > 0 {
		for _, k := range l.config.extraKeys {
			if l.config.extraKeyAsStr {
				v := ctx.Value(string(k))
				if v != nil {
					zapLogger = zapLogger.With(zap.Any(string(k), v))
				}
			} else {
				v := ctx.Value(k)
				if v != nil {
					zapLogger = zapLogger.With(zap.Any(string(k), v))
				}
			}
		}
	}
	log := zapLogger.Sugar()
	switch level {
	case hlog.LevelDebug, hlog.LevelTrace:
		log.Debugf(format, kvs...)
	case hlog.LevelInfo:
		log.Infof(format, kvs...)
	case hlog.LevelNotice, hlog.LevelWarn:
		log.Warnf(format, kvs...)
	case hlog.LevelError:
		log.Errorf(format, kvs...)
	case hlog.LevelFatal:
		log.Fatalf(format, kvs...)
	default:
		log.Warnf(format, kvs...)
	}
}

func (l *Logger) CtxLogf(level hlog.Level, ctx context.Context, format string, kvs ...interface{}) {
	var zLevel zapcore.Level
	zLevel = hLevelToZapLevel(level)
	if !l.config.coreConfigs[0].Lvl.Enabled(zLevel) {
		return
	}
	//获取调用者信息
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		pc, file, line, ok = runtime.Caller(1)
	}
	// 获取函数名, 命名空间
	var funcName string
	if fn := runtime.FuncForPC(pc); fn != nil {
		funcName = fn.Name()
	} else {
		funcName = "unknown"
	}
	var codeNamespace string
	// 提取命名空间
	codeNamespace = getDirectory(file, "/")
	// 创建新的 context 包含代码位置信息
	ctx = context.WithValue(ctx, codeFilepathKey, file)
	ctx = context.WithValue(ctx, codeFunctionKey, funcName)
	ctx = context.WithValue(ctx, codeLinenoKey, line)
	ctx = context.WithValue(ctx, codeNamespaceKey, codeNamespace)
	span := trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()
	traceFlags := span.SpanContext().TraceFlags()
	if span.SpanContext().IsValid() {
		ctx = context.WithValue(ctx, ExtraKey(traceIDKey), traceID)
		ctx = context.WithValue(ctx, ExtraKey(spanIDKey), spanID)
		ctx = context.WithValue(ctx, ExtraKey(traceFlagsKey), traceFlags)
		l.ctxLogf(level, ctx, format, kvs...)
	} else {
		l.ctxLogf(level, ctx, format, kvs...)
	}
	if !span.IsRecording() {
		return
	}
	switch level {
	case hlog.LevelDebug, hlog.LevelTrace:
		zLevel = zap.DebugLevel
	case hlog.LevelInfo:
		zLevel = zap.InfoLevel
	case hlog.LevelNotice, hlog.LevelWarn:
		zLevel = zap.WarnLevel
	case hlog.LevelError:
		zLevel = zap.ErrorLevel
	case hlog.LevelFatal:
		zLevel = zap.FatalLevel
	default:
		zLevel = zap.WarnLevel
	}

	// 设置 span 状态
	if zLevel >= l.config.traceConfig.errorSpanLevel {
		msg := getMessage(format, kvs)
		span.SetStatus(codes.Error, "")
		span.RecordError(errors.New(msg), trace.WithStackTrace(l.config.traceConfig.recordStackTraceInSpan))
	}
}

func (l *Logger) Trace(v ...interface{}) {
	l.Log(hlog.LevelTrace, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.Log(hlog.LevelDebug, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.Log(hlog.LevelInfo, v...)
}

func (l *Logger) Notice(v ...interface{}) {
	l.Log(hlog.LevelNotice, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.Log(hlog.LevelWarn, v...)
}

func (l *Logger) Error(v ...interface{}) {
	l.Log(hlog.LevelError, v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.Log(hlog.LevelFatal, v...)
}

func (l *Logger) Tracef(format string, v ...interface{}) {
	l.Logf(hlog.LevelTrace, format, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Logf(hlog.LevelDebug, format, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Logf(hlog.LevelInfo, format, v...)
}

func (l *Logger) Noticef(format string, v ...interface{}) {
	l.Logf(hlog.LevelWarn, format, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Logf(hlog.LevelWarn, format, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Logf(hlog.LevelError, format, v...)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Logf(hlog.LevelFatal, format, v...)
}

func (l *Logger) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelDebug, ctx, format, v...)
}

func (l *Logger) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelInfo, ctx, format, v...)
}

func (l *Logger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelWarn, ctx, format, v...)
}

func (l *Logger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelWarn, ctx, format, v...)
}

func (l *Logger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelError, ctx, format, v...)
}

func (l *Logger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	l.CtxLogf(hlog.LevelFatal, ctx, format, v...)
}

func (l *Logger) SetLevel(level hlog.Level) {
	lvl := hLevelToZapLevel(level)

	l.config.coreConfigs[0].Lvl = lvl

	cores := make([]zapcore.Core, 0, len(l.config.coreConfigs))
	for _, coreConfig := range l.config.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}

	logger := zap.New(
		zapcore.NewTee(cores[:]...),
		l.config.zapOpts...)

	l.l = logger
}

func (l *Logger) SetOutput(writer io.Writer) {
	l.config.coreConfigs[0].Ws = zapcore.AddSync(writer)

	cores := make([]zapcore.Core, 0, len(l.config.coreConfigs))
	for _, coreConfig := range l.config.coreConfigs {
		cores = append(cores, zapcore.NewCore(coreConfig.Enc, coreConfig.Ws, coreConfig.Lvl))
	}

	logger := zap.New(
		zapcore.NewTee(cores[:]...),
		l.config.zapOpts...)

	l.l = logger
}

// Logger is used to return an instance of *zap.Logger for custom fields, etc.
func (l *Logger) Logger() *zap.Logger {
	return l.l
}

func (l *Logger) Sync() {
	_ = l.l.Sync()
}
