// Package provider 定义翻译服务的统一抽象。
//
// 本文件是 SPEC-A（A1）的实现起点：接口与事件类型在此定稿（对齐
// packages/contracts/src/index.ts），实现方不得修改签名，只填充实现。
// 设计取向：回调式流、零产出才重试、截断防御；不含工具调用/reasoning/
// 多模态等翻译用不到的能力。
package provider

import "context"

// TextReq 单服务翻译请求（bridge 层的 TextReq 去掉 serviceIds 后逐服务下发）。
type TextReq struct {
	Text       string `json:"text"`
	SourceLang string `json:"sourceLang"` // "auto" 或 BCP-47 简化代码
	TargetLang string `json:"targetLang"`
}

// StreamEventType 流事件类型，契约见 packages/contracts README「流式事件时序」。
type StreamEventType string

const (
	EventMessageStart StreamEventType = "message-start"
	EventDelta        StreamEventType = "delta"
	EventStreamRetry  StreamEventType = "stream-retry"
	EventFinish       StreamEventType = "finish"
	EventError        StreamEventType = "error"
)

// ErrorCode 终态错误码（对应 contracts.StreamErrorCode）。
type ErrorCode string

const (
	CodeAuth            ErrorCode = "auth"
	CodeRateLimit       ErrorCode = "rate_limit"
	CodeTimeout         ErrorCode = "timeout"
	CodeConnect         ErrorCode = "connect"
	CodeContextOverflow ErrorCode = "context_overflow"
	CodeNetwork         ErrorCode = "network"
	CodeCancelled       ErrorCode = "cancelled"
	CodeUnknown         ErrorCode = "unknown"
)

// Usage token 用量（可选上报）。
type Usage struct {
	PromptTokens     int `json:"promptTokens,omitempty"`
	CompletionTokens int `json:"completionTokens,omitempty"`
}

// StreamEvent 流事件。RequestID/ServiceID 由 bridge 层填充，provider 实现只关心
// type 及载荷；bridge 负责经 Wails 事件 translate:event 推给前端。
type StreamEvent struct {
	RequestID string          `json:"requestId"`
	ServiceID string          `json:"serviceId"`
	Type      StreamEventType `json:"type"`
	Text      string          `json:"text,omitempty"`
	Message   string          `json:"message,omitempty"`
	ErrorCode ErrorCode       `json:"errorCode,omitempty"`
	Attempt   int             `json:"attempt,omitempty"`
	Usage     *Usage          `json:"usage,omitempty"`
}

// ProviderConfig 单个服务的运行时配置（从 store 的 services.config JSON 解出）。
type ProviderConfig struct {
	Type           string            `json:"type"` // llm | deepl | google | microsoft
	Name           string            `json:"name"`
	BaseURL        string            `json:"baseUrl,omitempty"`
	APIKey         string            `json:"apiKey,omitempty"`
	Model          string            `json:"model,omitempty"`
	AuthScheme     string            `json:"authScheme,omitempty"` // bearer（缺省）| x-api-key
	CustomHeaders  map[string]string `json:"customHeaders,omitempty"`
	SourceLangHint string            `json:"-"` // 内部用
}

// Provider 翻译服务适配器。只有流式：非流式服务（deepl/google/microsoft）
// 在内部完成请求后一次性 emit delta + finish。
//
// 约定：
//   - emit 返回 error 时立即终止流（消费方断开）
//   - ctx 取消时断开底层 HTTP 并返回 ctx.Err()
//   - 正常路径必须以 finish 或 error 事件收尾；干净 EOF 但未见服务端
//     终态标记视为断流（截断防御），交由重试层处理
type Provider interface {
	// TranslateStream 执行一次翻译并逐事件回调。emit 不得阻塞超过事件处理
	// 本身；重试（零产出才重试）由调用方 StreamWithRetry 统一包裹。
	TranslateStream(ctx context.Context, req TextReq, emit func(*StreamEvent) error) error
	// Name 服务协议标识（日志用）。
	Name() string
}
