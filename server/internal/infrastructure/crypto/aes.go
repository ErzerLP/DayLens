// Package crypto 提供 AES-256-GCM 加解密，用于保护敏感字段。
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// 加密前缀，用于区分加密数据和明文旧数据
const encPrefix = "ENC:"

// Cipher AES-256-GCM 加解密器
type Cipher struct {
	gcm cipher.AEAD
}

// NewCipher 从 32 字节 hex key 创建加密器
// key 长度必须为 16/24/32 字节（AES-128/192/256）
func NewCipher(key []byte) (*Cipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 16/24/32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("crypto: new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("crypto: new gcm: %w", err)
	}
	return &Cipher{gcm: gcm}, nil
}

// Encrypt 加密明文，返回 "ENC:<base64(nonce+ciphertext)>"
func (c *Cipher) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: read nonce: %w", err)
	}
	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文。如果数据没有 ENC: 前缀，视为旧明文数据直接返回。
func (c *Cipher) Decrypt(data string) (string, error) {
	if data == "" {
		return "", nil
	}
	// 兼容旧数据：没有前缀的是明文
	if !strings.HasPrefix(data, encPrefix) {
		return data, nil
	}

	raw, err := base64.StdEncoding.DecodeString(data[len(encPrefix):])
	if err != nil {
		return "", fmt.Errorf("crypto: base64 decode: %w", err)
	}

	nonceSize := c.gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", errors.New("crypto: ciphertext too short")
	}

	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: decrypt: %w", err)
	}
	return string(plaintext), nil
}

// EncryptPtr 加密 *string 指针
func (c *Cipher) EncryptPtr(p *string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	enc, err := c.Encrypt(*p)
	if err != nil {
		return nil, err
	}
	return &enc, nil
}

// DecryptPtr 解密 *string 指针
func (c *Cipher) DecryptPtr(p *string) (*string, error) {
	if p == nil {
		return nil, nil
	}
	dec, err := c.Decrypt(*p)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}

// NopCipher 无操作加密器（未配置 key 时使用，明文存储）
type NopCipher struct{}

func (NopCipher) Encrypt(plaintext string) (string, error)   { return plaintext, nil }
func (NopCipher) Decrypt(data string) (string, error)        { return data, nil }
func (NopCipher) EncryptPtr(p *string) (*string, error)      { return p, nil }
func (NopCipher) DecryptPtr(p *string) (*string, error)      { return p, nil }

// FieldCipher 字段加解密接口
type FieldCipher interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(data string) (string, error)
	EncryptPtr(p *string) (*string, error)
	DecryptPtr(p *string) (*string, error)
}
