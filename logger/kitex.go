package logger

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/klog"
)

// SetKitexLogger 把统一 logger 适配并安装为 kitex klog 的全局实现，
// 使 kitex 框架日志与业务日志走同一通道（zap + trace 注入）。
//
// hlog.Level 与 klog.Level 同源（均为 int 枚举且顺序一致），仅做数值转换。
func SetKitexLogger(lg *Logger) {
	klog.SetLogger(&kitexLogger{Logger: lg})
}

// kitexLogger 内嵌 *Logger 继承全部同签名方法（Debug/Info/...f/v/Ctx*f），
// 仅重写与 hlog 类型不同的四个方法以满足 klog.FullLogger。
type kitexLogger struct {
	*Logger
}

var _ klog.FullLogger = (*kitexLogger)(nil)

func (k *kitexLogger) Log(level klog.Level, kvs ...interface{}) {
	k.Logger.Log(hlog.Level(level), kvs...)
}

func (k *kitexLogger) Logf(level klog.Level, format string, kvs ...interface{}) {
	k.Logger.Logf(hlog.Level(level), format, kvs...)
}

func (k *kitexLogger) CtxLogf(level klog.Level, ctx context.Context, format string, kvs ...interface{}) {
	k.Logger.CtxLogf(hlog.Level(level), ctx, format, kvs...)
}

func (k *kitexLogger) SetLevel(level klog.Level) {
	k.Logger.SetLevel(hlog.Level(level))
}
