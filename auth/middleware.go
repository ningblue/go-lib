package auth

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/ningblue/go-lib/render"
)

// Options 配置认证中间件。
type Options struct {
	// Secret 为 HS256 共享密钥（必填）。
	Secret []byte
	// Optional 为 true 时不强制认证：无 token 直接放行，有 token 则解析注入。
	Optional bool
	// HeaderName 自定义承载 token 的请求头，默认 Authorization。
	HeaderName string
	// QueryParam 允许从该 query 参数取 token（WebSocket / SSE 场景），默认 token。
	QueryParam string
}

// Middleware 校验 access token 并注入操作者上下文。
// 鉴权失败渲染 401422（业务码）后中止链；成功则 c.Next。
func Middleware(opt Options) app.HandlerFunc {
	headerName := opt.HeaderName
	if headerName == "" {
		headerName = "Authorization"
	}
	queryParam := opt.QueryParam
	if queryParam == "" {
		queryParam = "token"
	}
	return func(ctx context.Context, c *app.RequestContext) {
		raw := extractToken(c, headerName, queryParam)
		if raw == "" {
			if opt.Optional {
				c.Next(ctx)
				return
			}
			render.Unauthorized(c, "missing access token")
			c.Abort()
			return
		}
		claims, err := Parse(raw, opt.Secret, TokenTypeAccess)
		if err != nil {
			if opt.Optional {
				c.Next(ctx)
				return
			}
			render.Unauthorized(c, "invalid access token")
			c.Abort()
			return
		}
		SetOperator(c, Operator{
			UserID:   claims.UserID,
			UserKey:  claims.UserKey,
			Username: claims.Username,
			TenantID: claims.TenantID,
			Role:     claims.Role,
		})
		c.Next(ctx)
	}
}

// extractToken 依次从 Authorization 头（Bearer）与 query 参数取 token。
func extractToken(c *app.RequestContext, headerName, queryParam string) string {
	if v := c.GetHeader(headerName); len(v) > 0 {
		s := string(v)
		if i := strings.IndexByte(s, ' '); i >= 0 {
			s = s[i+1:]
		}
		if s = strings.TrimSpace(s); s != "" {
			return s
		}
	}
	return strings.TrimSpace(c.Query(queryParam))
}

// RequireRole 在认证中间件之后使用：限定可访问角色（大小写不敏感），不匹配渲染 403。
// roles 为空表示不限制。
func RequireRole(roles ...string) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if len(roles) == 0 {
			c.Next(ctx)
			return
		}
		op, ok := OperatorFrom(c)
		if !ok {
			render.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}
		for _, r := range roles {
			if strings.EqualFold(op.Role, r) {
				c.Next(ctx)
				return
			}
		}
		render.Forbidden(c, "permission denied")
		c.Abort()
	}
}
