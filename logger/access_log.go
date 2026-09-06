package logger

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// AccessLog 返回 hertz 访问日志中间件：结构化输出每个请求的
// method/path/status/latency/client_ip/user_agent。
//
// 约定：
//   - 输出通道与级别由全局统一 logger 决定（先 hlog.SetLogger 设定 go-lib logger），
//     本中间件不单独建文件/目录；
//   - trace_id/span_id 由统一 logger 从 ctx 自动注入（enableTrace 时），
//     因此本中间件应挂在 tracing 中间件之内（后注册）。
func AccessLog() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		hlog.CtxInfof(ctx, "access %s %s status=%d latency=%s ip=%s ua=%q",
			string(c.Method()), string(c.Path()),
			c.Response.StatusCode(),
			time.Since(start).Round(time.Millisecond).String(),
			c.ClientIP(), string(c.UserAgent()),
		)
	}
}
