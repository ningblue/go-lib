// Package auth 提供跨服务的 JWT 会话与操作者上下文（身份传递）基础能力。
// 业务角色/权限模型不在此包（属各平台业务），本包只做机制：签发、校验、上下文存取。
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token 类型。access 用于访问，refresh 仅用于换发 access。
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Claims 是本平台统一 JWT 声明。tenant_id 为空表示平台级（super_admin 全局身份）。
type Claims struct {
	UserID    string `json:"userId"`
	UserKey   string `json:"userKey,omitempty"`
	Username  string `json:"username"`
	TenantID  string `json:"tenantId"`
	Role      string `json:"role"`
	TokenType string `json:"tokenType"`
	jwt.RegisteredClaims
}

// GenerateOption 签发参数。
type GenerateOption struct {
	// Secret 为 HS256 共享密钥，必填。
	Secret []byte
	// TTL 为有效期，<=0 时用 DefaultAccessTTL（access）/DefaultRefreshTTL（refresh）。
	TTL time.Duration
}

// 默认有效期（与老栈一致：access 24h / refresh 7d）。
const (
	DefaultAccessTTL  = 24 * time.Hour
	DefaultRefreshTTL = 7 * 24 * time.Hour
)

var (
	// ErrInvalidToken 表示 token 缺失、格式错误、签名不匹配或已过期。
	ErrInvalidToken = errors.New("invalid token")
	// ErrWrongTokenType 表示 token 类型不符合调用方期望（如用 refresh 当 access）。
	ErrWrongTokenType = errors.New("unexpected token type")
)

func ttlFor(tokenType string, ttl time.Duration) time.Duration {
	if ttl > 0 {
		return ttl
	}
	if tokenType == TokenTypeRefresh {
		return DefaultRefreshTTL
	}
	return DefaultAccessTTL
}

// Generate 按 claims 与 tokenType 签发 HS256 token。
func Generate(c Claims, tokenType string, opt GenerateOption) (string, error) {
	if len(opt.Secret) == 0 {
		return "", errors.New("auth: jwt secret is empty")
	}
	now := time.Now()
	c.TokenType = tokenType
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(ttlFor(tokenType, opt.TTL))),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString(opt.Secret)
}

// Parse 校验签名与有效期并返回 claims；expectedType 非空时还要求 tokenType 匹配。
func Parse(tokenString string, secret []byte, expectedType string) (*Claims, error) {
	if tokenString == "" || len(secret) == 0 {
		return nil, ErrInvalidToken
	}
	claims := &Claims{}
	tok, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !tok.Valid {
		return nil, ErrInvalidToken
	}
	if expectedType != "" && claims.TokenType != expectedType {
		return nil, ErrWrongTokenType
	}
	return claims, nil
}

// ExpiresIn 返回 access token 的剩余秒数（<=0 表示已过期）；用于响应体 expires_in。
func (c *Claims) ExpiresIn() int64 {
	if c == nil || c.ExpiresAt == nil {
		return 0
	}
	d := int64(time.Until(c.ExpiresAt.Time).Seconds())
	if d < 0 {
		return 0
	}
	return d
}
