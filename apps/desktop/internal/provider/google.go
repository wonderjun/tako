// Google 网页免费接口适配器：GET translate_a/single?client=gtx（免 key）。
// 非流式：完成后一次性 emit message-start → delta → finish。
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// googleTranslateBaseURL 网页免费接口（client=gtx 路由无需 key/tk 参数）。
const googleTranslateBaseURL = "https://translate.googleapis.com"

// GoogleProvider Google 免费翻译适配器。
type GoogleProvider struct {
	baseURL string
	client  *http.Client
}

// NewGoogle 构造 Google 适配器（无需 apiKey）。
func NewGoogle(cfg ProviderConfig) (*GoogleProvider, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = googleTranslateBaseURL
	}
	return &GoogleProvider{baseURL: base, client: NewHTTPClient(30 * time.Second)}, nil
}

// Name 服务协议标识。
func (p *GoogleProvider) Name() string { return "google" }

// googleLangMap BCP-47 简化代码 → Google 代码；未列出的直接透传。
var googleLangMap = map[string]string{
	"zh-Hans": "zh-CN", "zh-Hant": "zh-TW",
}

func googleLang(code string) string {
	if code == "auto" || code == "" {
		return "auto"
	}
	if mapped, ok := googleLangMap[code]; ok {
		return mapped
	}
	return code
}

// TranslateStream 实现 Provider：非流式请求完成后一次性回放事件序列。
func (p *GoogleProvider) TranslateStream(
	ctx context.Context,
	req TextReq,
	emit func(*StreamEvent) error,
) error {
	q := url.Values{
		"client": {"gtx"},
		"sl":     {googleLang(req.SourceLang)},
		"tl":     {googleLang(req.TargetLang)},
		"dt":     {"t"},
		"q":      {req.Text},
	}
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		p.baseURL+"/translate_a/single?"+q.Encode(), nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Tako/1.0")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return NormalizeStreamError(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return NormalizeStreamError(err)
	}
	if resp.StatusCode >= 400 {
		return &ProviderError{
			Code:       Classify(resp.StatusCode, string(raw)),
			StatusCode: resp.StatusCode,
			Detail:     ExtractErrorDetail(string(raw), 300),
		}
	}

	// 响应形态：[[["译文","原文",...],...],null,"en",...]——取第 0 段每条的
	// 第 0 元素拼接为完整译文。
	var parsed []any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return &ProviderError{Code: CodeUnknown, Detail: "google 响应解析失败: " + err.Error(), Cause: err}
	}
	var b strings.Builder
	if segs, ok := anySliceIndex(parsed, 0); ok {
		for _, seg := range segs {
			if arr, ok := seg.([]any); ok && len(arr) > 0 {
				if s, ok := arr[0].(string); ok {
					b.WriteString(s)
				}
			}
		}
	}
	text := b.String()
	if text == "" {
		return &ProviderError{Code: CodeUnknown, Detail: fmt.Sprintf("google 响应无译文: %s", Truncate(string(raw), 200))}
	}
	return emitOnce(emit, text)
}

// anySliceIndex 安全取嵌套数组元素。
func anySliceIndex(v []any, i int) ([]any, bool) {
	if i >= len(v) {
		return nil, false
	}
	seg, ok := v[i].([]any)
	return seg, ok
}
