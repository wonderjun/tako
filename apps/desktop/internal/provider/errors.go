// 错误归一化：HTTP 状态码/响应体/底层网络错误 → contracts 的 8 个
// StreamErrorCode。分类结果决定两件事：终态 error 事件的 errorCode 字段，
// 以及 StreamWithRetry 的「可重试」判定（瞬时错误才重试）。
package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
)

// ProviderError 归一化的 provider 错误。Code 对齐 contracts.StreamErrorCode；
// StatusCode 为上游 HTTP 状态码（无则 0）；Detail 是面向用户的脱敏详情；
// Cause 保持底层错误链（net.Error / context 错误可经 errors.Is/As 穿透）。
type ProviderError struct {
	Code       ErrorCode
	StatusCode int
	Detail     string
	Cause      error
}

func (e *ProviderError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("provider error (%s, http %d): %s", e.Code, e.StatusCode, e.Detail)
	}
	return fmt.Sprintf("provider error (%s): %s", e.Code, e.Detail)
}

// Unwrap 保持底层错误可穿透（net.Error 超时判定、context.Canceled 等）。
func (e *ProviderError) Unwrap() error { return e.Cause }

// contextOverflowPatterns 上下文窗口溢出特征串（小写包含匹配）。
var contextOverflowPatterns = []string{
	"context_length_exceeded",
	"context length",
	"context window",
	"maximum context",
	"context overflow",
	"prompt is too long",
	"too many tokens",
	"exceed context limit",
	"exceeds the context",
	"reduce the length",
	"token limit exceeded",
}

// Classify 由上游 HTTP 状态码 + 响应体归一出错误码（各适配器共用）。
//
// 判定顺序：溢出特征串优先（400 伴随溢出文案应归 context_overflow 而非笼统
// unknown）→ 明确状态码 → 通用文案特征 → 5xx 兜底 network。
func Classify(statusCode int, body string) ErrorCode {
	lower := strings.ToLower(body)
	for _, pattern := range contextOverflowPatterns {
		if strings.Contains(lower, pattern) {
			return CodeContextOverflow
		}
	}
	switch statusCode {
	case 401, 403:
		return CodeAuth
	case 408:
		return CodeTimeout
	case 429:
		return CodeRateLimit
	}
	if strings.Contains(lower, "rate_limit") || strings.Contains(lower, "rate limit") ||
		strings.Contains(lower, "too many requests") {
		return CodeRateLimit
	}
	if strings.Contains(lower, "authentication") || strings.Contains(lower, "unauthorized") ||
		strings.Contains(lower, "invalid api key") || strings.Contains(lower, "invalid_api_key") {
		return CodeAuth
	}
	if statusCode >= 500 {
		// 5xx（500/502/503/504/529 过载等）统一归 network：上游故障，零产出可重试。
		return CodeNetwork
	}
	return CodeUnknown
}

// ErrorCodeOf 沿错误链取首个 *ProviderError 的 Code；取不到时按底层错误
// 形态兜底（context 取消 → cancelled，net 超时 → timeout，其余 → unknown）。
func ErrorCodeOf(err error) ErrorCode {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe.Code
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return CodeCancelled
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return CodeTimeout
	}
	return CodeUnknown
}

// IsTransient 判断是否可自动重试的瞬时错误：限流/超时/传输层断连/上游 5xx。
// 鉴权失败、上下文溢出、用户取消、未知错误一律不重试（重试也只会同样失败，
// 或造成重复计费的语义错误）。
func IsTransient(err error) bool {
	switch ErrorCodeOf(err) {
	case CodeRateLimit, CodeTimeout, CodeConnect, CodeNetwork:
		return true
	}
	return false
}

// ExtractErrorDetail 从上游响应体提取用户可读的错误详情：依次尝试 JSON 的
// error.message / message / error（字符串），解析失败回退原文，统一截断。
func ExtractErrorDetail(body string, limit int) string {
	text := strings.TrimSpace(body)
	if text == "" {
		return ""
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return Truncate(text, limit)
	}
	if data == nil {
		return Truncate(text, limit)
	}
	if errorObj, ok := data["error"].(map[string]any); ok {
		if msg, ok := errorObj["message"].(string); ok && strings.TrimSpace(msg) != "" {
			return Truncate(strings.TrimSpace(msg), limit)
		}
	}
	if msg, ok := data["message"].(string); ok {
		return Truncate(strings.TrimSpace(msg), limit)
	}
	if errStr, ok := data["error"].(string); ok && strings.TrimSpace(errStr) != "" {
		return Truncate(strings.TrimSpace(errStr), limit)
	}
	return Truncate(text, limit)
}

// Truncate 按字符数截断（limit<=0 不截断）。供事件 message 与错误详情共用。
func Truncate(s string, limit int) string {
	if limit > 0 && len(s) > limit {
		return s[:limit]
	}
	return s
}

// 传输层错误归一化的判定辅助（httpx.go 与 llm/openai.go 共用）。

// IsConnectError 仅匹配连接建立（dial）阶段失败：等响应头/读正文阶段的
// 错误不得触发连接重试（避免请求未送达语义之外的误判）。
func IsConnectError(err error) bool {
	var opErr *net.OpError
	return errors.As(err, &opErr) && opErr.Op == "dial"
}

// IsStreamBreakError 判断读流阶段的传输层断连：reset/broken pipe 等非超时
// net.OpError，以及帧中途截断的 unexpected EOF。
func IsStreamBreakError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) && !opErr.Timeout() {
		return true
	}
	return errors.Is(err, io.ErrUnexpectedEOF)
}
