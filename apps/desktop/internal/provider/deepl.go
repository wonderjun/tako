// DeepL Free API 适配器：POST /v2/translate（api-free.deepl.com），
// 鉴权头 Authorization: DeepL-Auth-Key。非流式：完成后一次性 emit
// message-start → delta → finish。
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

// deepLFreeBaseURL DeepL Free API 默认接入点（Pro 为 api.deepl.com，
// 可由 ProviderConfig.BaseURL 覆盖）。
const deepLFreeBaseURL = "https://api-free.deepl.com"

// DeepLProvider DeepL 翻译适配器。
type DeepLProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewDeepL 构造 DeepL 适配器。apiKey 必填（缺省视为鉴权配置缺失）。
func NewDeepL(cfg ProviderConfig) (*DeepLProvider, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, &ProviderError{Code: CodeAuth, Detail: "deepl 服务缺少 apiKey"}
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = deepLFreeBaseURL
	}
	return &DeepLProvider{
		baseURL: base,
		apiKey:  cfg.APIKey,
		client:  NewHTTPClient(30 * time.Second),
	}, nil
}

// Name 服务协议标识。
func (p *DeepLProvider) Name() string { return "deepl" }

// deepLLangMap BCP-47 简化代码 → DeepL 代码。source_lang 侧 DeepL 只认 ZH
// （简体），target_lang 侧 zh-Hant 映射 ZH-HANT。
var deepLLangMap = map[string]string{
	"zh-Hans": "ZH", "zh-Hant": "ZH-HANT", "en": "EN", "ja": "JA", "ko": "KO",
	"fr": "FR", "de": "DE", "es": "ES", "pt": "PT", "it": "IT", "ru": "RU",
	"ar": "AR", "th": "TH", "vi": "VI", "id": "ID",
}

func deepLLang(code string, forSource bool) string {
	mapped, ok := deepLLangMap[code]
	if !ok {
		return ""
	}
	// source_lang 不支持 ZH-HANT（DeepL 源侧只有 ZH）。
	if forSource && mapped == "ZH-HANT" {
		return "ZH"
	}
	return mapped
}

// TranslateStream 实现 Provider：非流式请求完成后一次性回放事件序列。
func (p *DeepLProvider) TranslateStream(
	ctx context.Context,
	req TextReq,
	emit func(*StreamEvent) error,
) error {
	form := url.Values{
		"text":        {req.Text},
		"target_lang": {deepLLang(req.TargetLang, false)},
	}
	if src := deepLLang(req.SourceLang, true); req.SourceLang != "auto" && src != "" {
		form.Set("source_lang", src)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.baseURL+"/v2/translate", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "DeepL-Auth-Key "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		code := Classify(resp.StatusCode, string(raw))
		// DeepL 配额/限流是 456（QUOTA_EXCEEDED / TOO_MANY_REQUESTS 语义）。
		if resp.StatusCode == 456 && code == CodeUnknown {
			code = CodeRateLimit
		}
		return &ProviderError{
			Code:       code,
			StatusCode: resp.StatusCode,
			Detail:     ExtractErrorDetail(string(raw), 300),
		}
	}

	var parsed struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return &ProviderError{Code: CodeUnknown, Detail: "deepl 响应解析失败: " + err.Error(), Cause: err}
	}
	if len(parsed.Translations) == 0 {
		return &ProviderError{Code: CodeUnknown, Detail: fmt.Sprintf("deepl 响应无 translations: %s", Truncate(string(raw), 200))}
	}
	return emitOnce(emit, parsed.Translations[0].Text)
}
