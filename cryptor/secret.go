// 密钥加解密便捷层：AES-256-GCM + base64 文本承载。
// 存储格式与老栈一致（nonce 前缀 + GCM 密文，base64 标准编码），便于凭据数据迁移。
package cryptor

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// ErrSecretKeySize 表示密钥长度不合法。
var ErrSecretKeySize = errors.New("cryptor: secret key must be 32 bytes (raw, 64-char hex, or base64)")

// ParseSecretKey 解析 32 字节密钥：支持原始 32 字节、64 字符 hex、base64 编码 32 字节。
func ParseSecretKey(raw string) ([]byte, error) {
	if len(raw) == 32 {
		return []byte(raw), nil
	}
	if decoded, err := hex.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	if decoded, err := base64.StdEncoding.DecodeString(raw); err == nil && len(decoded) == 32 {
		return decoded, nil
	}
	return nil, ErrSecretKeySize
}

// EncryptSecret 以 key 加密明文，返回 base64 文本（nonce 前缀）。
func EncryptSecret(plaintext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrSecretKeySize
	}
	sealed, err := AesGcmEncrypt([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptSecret 解密 EncryptSecret 产物；密文损坏/密钥不符返回错误（不 panic）。
func DecryptSecret(ciphertext string, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrSecretKeySize
	}
	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("cryptor: ciphertext is invalid")
	}
	nonce, encrypted := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
