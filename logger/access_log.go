package logger

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"time"
)

// RequestIDHeaderValue value for the request id header
const RequestIDHeaderValue = "X-Request-ID"

// LoggerMiddleware middleware for logging incoming requests
func loggerMiddleware(level string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		// 可定制的输出目录。
		logFilePath := "./logs/"
		if err := os.MkdirAll(logFilePath, 0o777); err != nil {
			hlog.Error("Failed to create log directory", zap.Error(err))
			return
		}
		// 将文件名设置为日期
		logFileName := "access.log"
		fileName := path.Join(logFilePath, logFileName)
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			if _, err := os.Create(fileName); err != nil {
				hlog.Error("Failed to create log file", zap.Error(err))
				return
			}
		}
		dynamicLevel, err := zap.ParseAtomicLevel(level)
		if err != nil {
			hlog.Error("parse log level error", err)
			return
		}

		file := zapcore.NewCore(
			zapcore.NewJSONEncoder(humanEncoderConfig()),
			zapcore.AddSync(getWriteSyncer(fileName)),
			dynamicLevel,
		)
		logger := zap.New(file)
		c.Set("logger", logger)
		if reqId, ok := ctx.Value(RequestIDHeaderValue).(string); ok {
			//logger = logger.With(zap.String("request_id", reqId))
			logger.With(zap.String("request_id", reqId))
		}
		defer func() {
			stop := time.Now()
			logger.Info("request processed",
				zap.String("datetime", time.Now().Format("2006-01-02 15:04:05")),
				zap.String("remote_ip", c.ClientIP()),
				zap.String("method", string(c.Method())),
				zap.String("path", string(c.Path())),
				zap.Int("status", c.Response.StatusCode()),
				zap.Duration("latency", stop.Sub(start)),
				zap.String("latency_human", stop.Sub(start).String()),
				zap.String("user_agent", string(c.UserAgent())),
			)
		}()
		c.Next(ctx)
	}
}

func InitAccessLogger(level string) app.HandlerFunc {
	return loggerMiddleware(level)
}
