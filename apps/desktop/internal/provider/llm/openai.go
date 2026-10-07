// Package llm LLM 翻译适配器：OpenAI 兼容协议（chat/completions SSE）。
// 只保留 chat_completions 单格式，不含 responses/tools/reasoning/providerRaw
// 等翻译用不到的能力。
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"tako/internal/provider"
)

// OpenAIProvider OpenAI 兼容适配器（POST {baseURL}/chat/completions，SSE 流式）。
type OpenAIProvider struct {
	baseURL       string
	apiKey        string
	model         string
	authScheme    string // bearer（缺省）| x-api-key
	customHeaders map[string]string
	client        *http.Client
}

// New 由运行时配置构造适配器。baseUrl 与 model 必填；authScheme 缺省 bearer。
func New(cfg provider.ProviderConfig) (*OpenAIProvider, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		return nil, &provider.ProviderError{Code: provider.CodeUnknown, Detail: "llm 服务缺少 baseUrl"}
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, &provider.ProviderError{Code: provider.CodeUnknown, Detail: "llm 服务缺少 model"}
	}
	scheme := strings.ToLower(strings.TrimSpace(cfg.AuthScheme))
	if scheme == "" {
		scheme = "bearer"
	}
	if scheme != "bearer" && scheme != "x-api-key" {
		return nil, &provider.ProviderError{
			Code:   provider.CodeUnknown,
			Detail: fmt.Sprintf("不支持的鉴权方案: %s", cfg.AuthScheme),
		}
	}
	headers := make(map[string]string, len(cfg.CustomHeaders))
	for k, v := range cfg.CustomHeaders {
		headers[k] = v
	}
	return &OpenAIProvider{
		baseURL:       baseURL,
		apiKey:        cfg.APIKey,
		model:         strings.TrimSpace(cfg.Model),
		authScheme:    scheme,
		customHeaders: headers,
		client:        provider.NewStreamHTTPClient(),
	}, nil
}

// Name 服务协议标识（Provider 接口）。
func (p *OpenAIProvider) Name() string { return "llm" }

// TranslateStream 执行一次流式翻译：message-start → delta* → finish。
// 干净 EOF 但未见终态标记（[DONE]/finish_reason）视为断流，返回 connect
// 错误交给重试层裁决（零产出才重试）。
func (p *OpenAIProvider) TranslateStream(
	ctx context.Context,
	req provider.TextReq,
	emit func(*provider.StreamEvent) error,
) error {
	payload := map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": buildSystemPrompt(req)},
			{"role": "user", "content": req.Text},
		},
		"stream":         true,
		"stream_options": map[string]any{"include_usage": true},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return &provider.ProviderError{Code: provider.CodeUnknown, Detail: "请求体序列化失败: " + err.Error(), Cause: err}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	for k, v := range p.buildHeaders() {
		httpReq.Header.Set(k, v)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return provider.NormalizeStreamError(err)
	}
	body := provider.NewIdleWatchdogReadCloser(resp.Body, 0)
	defer body.Close()

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(io.LimitReader(body, 64*1024))
		detail := provider.ExtractErrorDetail(string(raw), 300)
		return &provider.ProviderError{
			Code:       provider.Classify(resp.StatusCode, string(raw)),
			StatusCode: resp.StatusCode,
			Detail:     detail,
		}
	}

	// 上游声明的终态标记：[DONE] 或带 finish_reason 的收尾块。
	// 干净 EOF 若缺终态标记 = 传输层断流，已收内容只是残片。
	sawTerminal := false
	started := false
	var full strings.Builder
	var usage *provider.Usage

	scanner := bufio.NewScanner(body)
	// SSE 单行上限放宽到 16MB（个别接入点单行 chunk 很大）
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		dataStr := strings.TrimSpace(line[len("data:"):])
		if dataStr == "[DONE]" {
			sawTerminal = true
			break
		}

		var chunk map[string]any
		if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
			continue
		}

		// 流内错误检测（任意位置）：部分接入点用无 choices 的 error 对象作
		// 首包，也有在内容输出中途报错的——静默吞掉会表现为"截断却正常收尾"。
		if rawErr, has := chunk["error"]; has && rawErr != nil {
			detail := provider.ExtractErrorDetail(dataStr, 300)
			return &provider.ProviderError{
				Code:   provider.Classify(0, dataStr),
				Detail: detail,
			}
		}

		if !started {
			started = true
			if err := emit(&provider.StreamEvent{Type: provider.EventMessageStart}); err != nil {
				return err
			}
		}

		if u := parseUsage(chunk["usage"]); u != nil {
			usage = u
		}

		choices, _ := chunk["choices"].([]any)
		for _, ch := range choices {
			choice, ok := ch.(map[string]any)
			if !ok {
				continue
			}
			delta, _ := choice["delta"].(map[string]any)
			if content, ok := delta["content"].(string); ok && content != "" {
				full.WriteString(content)
				if err := emit(&provider.StreamEvent{
					Type: provider.EventDelta,
					Text: content,
				}); err != nil {
					return err
				}
			}
			if fr, ok := choice["finish_reason"].(string); ok && fr != "" {
				sawTerminal = true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return provider.NormalizeStreamError(err)
	}

	if !sawTerminal {
		return &provider.ProviderError{
			Code:   provider.CodeConnect,
			Detail: "流被上游提前截断（未见终态标记）",
		}
	}

	return emit(&provider.StreamEvent{
		Type:  provider.EventFinish,
		Text:  full.String(),
		Usage: usage,
	})
}

// buildHeaders customHeaders 先并入 → 剥离 authorization/x-api-key（防静态
// 配置伪造鉴权头）→ 由 apiKey 按 authScheme 重建。apiKey 为空不发鉴权头。
func (p *OpenAIProvider) buildHeaders() map[string]string {
	headers := map[string]string{"Content-Type": "application/json"}
	for k, v := range p.customHeaders {
		headers[k] = v
	}
	for k := range headers {
		lk := strings.ToLower(k)
		if lk == "authorization" || lk == "x-api-key" {
			delete(headers, k)
		}
	}
	if p.apiKey != "" {
		switch p.authScheme {
		case "x-api-key":
			headers["x-api-key"] = p.apiKey
		default:
			headers["Authorization"] = "Bearer " + p.apiKey
		}
	}
	return headers
}

// buildSystemPrompt 翻译系统提示词：只回译文，不解释。
func buildSystemPrompt(req provider.TextReq) string {
	var b strings.Builder
	b.WriteString("你是专业翻译引擎。")
	if req.SourceLang != "" && req.SourceLang != "auto" {
		fmt.Fprintf(&b, "将用户输入从 %s 翻译为 %s。", req.SourceLang, req.TargetLang)
	} else {
		fmt.Fprintf(&b, "自动检测用户输入的语言并翻译为 %s。", req.TargetLang)
	}
	b.WriteString("只输出译文，不要解释、不要加引号、不要复述原文。")
	return b.String()
}

// parseUsage 从 chunk.usage 提取 token 用量（OpenAI snake_case 字段）。
func parseUsage(v any) *provider.Usage {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	u := &provider.Usage{
		PromptTokens:     safeInt(m["prompt_tokens"]),
		CompletionTokens: safeInt(m["completion_tokens"]),
	}
	if u.PromptTokens == 0 && u.CompletionTokens == 0 {
		return nil
	}
	return u
}

func safeInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	}
	return 0
}
