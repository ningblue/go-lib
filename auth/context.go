package auth

import (
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// 操作者上下文的 RequestContext 键（跨服务统一约定；api-gateway 与本平台管理面共用）。
const (
	ctxKeyUserID   = "userId"
	ctxKeyUserKey  = "userKey"
	ctxKeyUsername = "username"
	ctxKeyTenantID = "tenantId"
	ctxKeyRole     = "role"
)

// Operator 是一次请求的已认证身份（由 JWT 中间件注入）。
type Operator struct {
	UserID   string
	UserKey  string
	Username string
	TenantID string
	Role     string
}

// SetOperator 把身份写入请求上下文（认证中间件调用）。
func SetOperator(c *app.RequestContext, op Operator) {
	c.Set(ctxKeyUserID, op.UserID)
	c.Set(ctxKeyUserKey, op.UserKey)
	c.Set(ctxKeyUsername, op.Username)
	c.Set(ctxKeyTenantID, op.TenantID)
	c.Set(ctxKeyRole, op.Role)
}

// OperatorFrom 读取请求上下文中的身份；未认证时返回 false。
func OperatorFrom(c *app.RequestContext) (Operator, bool) {
	uid := c.GetString(ctxKeyUsername)
	if uid == "" {
		return Operator{}, false
	}
	return Operator{
		UserID:   c.GetString(ctxKeyUserID),
		UserKey:  c.GetString(ctxKeyUserKey),
		Username: uid,
		TenantID: c.GetString(ctxKeyTenantID),
		Role:     c.GetString(ctxKeyRole),
	}, true
}

// IsSuperAdmin 判断角色是否平台超管（大小写不敏感，与老栈口径一致）。
func (o Operator) IsSuperAdmin() bool {
	return strings.EqualFold(o.Role, "super_admin")
}

// EffectiveTenantID 返回身份实际生效的租户：super_admin 为全局（空串），其余为自身租户。
func (o Operator) EffectiveTenantID() string {
	if o.IsSuperAdmin() {
		return ""
	}
	return strings.TrimSpace(o.TenantID)
}

// ResolveTenant 解析本次请求的目标租户：super_admin 可用 requested 指定任意租户（空则全局），
// 普通用户只能操作自身租户（请求越过自身租户返回 false，由调用方渲染 403）。
func (o Operator) ResolveTenant(requested string) (string, bool) {
	if o.IsSuperAdmin() {
		return strings.TrimSpace(requested), true
	}
	own := strings.TrimSpace(o.TenantID)
	if r := strings.TrimSpace(requested); r != "" && r != own {
		return "", false
	}
	return own, true
}
