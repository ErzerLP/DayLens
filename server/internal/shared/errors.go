// Package shared 定义跨层共享的错误码和哨兵错误。
package shared

import "errors"

// 哨兵错误 — 跨层使用 errors.Is() 判断
var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("duplicate entry")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrFieldMissing = errors.New("required field missing")
	ErrFieldInvalid = errors.New("invalid field value")
	ErrUnavailable  = errors.New("service unavailable")
)

// 错误码常量 — HTTP 响应使用
const (
	CodeSuccess      = 0
	CodeTokenMissing = 40001
	CodeTokenInvalid = 40002
	CodeForbidden    = 40003
	CodeBadRequest   = 42201
	CodeFieldMissing = 42202
	CodeFieldInvalid = 42203
	CodeNotFound     = 40401
	CodeDuplicate    = 40901
	CodeInternal     = 50001
	CodeDatabase     = 50002
	CodeAIService    = 50003
	CodeUnavailable  = 50301
)
