// Microsoft（Edge）免费翻译适配器：免用户 key——先从 Edge 换 JWT，
// 再调 cognitive.microsofttranslator.com/translate。非流式：完成后一次性
// emit message-start → delta → finish。
package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// edgeAuthURL Edge 免费 token 端点（GET 返回 JWT 纯文本）。
	edgeAuthURL = "https://edge.microsoft.com/translate/auth"
	// edgeTranslateURL Edge 翻译端点（免 key，走上面换的 JWT）。
	edgeTranslateURL = "https://api-edge.cognitive.microsofttranslator.com/translate"
)

// edgeTokenLeeway token 提前过期余量（距 exp 不足此值即重新换取）。
const edgeTokenLeeway = 60 * time.Second

// MicrosoftProvider Microsoft Edge 免费翻译适配器。
type MicrosoftProvider struct {
	authURL      string
	translateURL string
	client       *http.Client

	mu        sync.Mutex
	token     string
	tokenExp  time.Time
	hasExpiry bool
}

// NewMicrosoft 构造 Microsoft 适配器（免 key；BaseURL 可覆盖翻译端点）。
func NewMicrosoft(cfg ProviderConfig) (*MicrosoftProvider, error) {
	translateURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/") + "/translate"
	if strings.TrimSpace(cfg.BaseURL) == "" {
		translateURL = edgeTranslateURL
	}
	return &MicrosoftProvider{
		authURL:      edgeAuthURL,
		translateURL: translateURL,
		client:       NewHTTPClient(30 * time.Second),
	}, nil
}

// Name 服务协议标识。
func (p *MicrosoftProvider) Name() string { return "microsoft" }

// edgeLang Edge 语言码：BCP-47 简化代码基本直通（zh-Hans/zh-Hant 原生支持）。
func edgeLang(code string) string { return code }

// TranslateStream 实现 Provider：非流式请求完成后一次性回放事件序列。
// token 失效（401）时作废缓存重取一次再试。
func (p *MicrosoftProvider) TranslateStream(
	ctx context.Context,
	req TextReq,
	emit func(*StreamEvent) error,
) error {
	const maxAuthRetries = 1
	var lastErr error
	for attempt := 0; attempt <= maxAuthRetries; attempt++ {
		text, err := p.translateOnce(ctx, req, attempt > 0)
		if err == nil {
			return emitOnce(emit, text)
		}
		lastErr = err
		// 只有鉴权类失败才值得重取 token 再试一次。
		if ErrorCodeOf(err) != CodeAuth {
			break
		}
	}
	return lastErr
}

// translateOnce 单次翻译调用；forceRefreshToken 作废缓存 token 重取。
func (p *MicrosoftProvider) translateOnce(ctx context.Context, req TextReq, forceRefreshToken bool) (string, error) {
	token, err := p.getToken(ctx, forceRefreshToken)
	if err != nil {
		return "", err
	}

	q := url.Values{"api-version": {"3.0"}, "to": {edgeLang(req.TargetLang)}}
	if req.SourceLang != "" && req.SourceLang != "auto" {
		q.Set("from", edgeLang(req.SourceLang))
	}
	payload, _ := json.Marshal([]map[string]string{{"Text": req.Text}})

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		p.translateURL+"?"+q.Encode(), bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", NormalizeStreamError(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", NormalizeStreamError(err)
	}
	if resp.StatusCode >= 400 {
		return "", &ProviderError{
			Code:       Classify(resp.StatusCode, string(raw)),
			StatusCode: resp.StatusCode,
			Detail:     ExtractErrorDetail(string(raw), 300),
		}
	}

	var parsed []struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", &ProviderError{Code: CodeUnknown, Detail: "microsoft 响应解析失败: " + err.Error(), Cause: err}
	}
	if len(parsed) == 0 || len(parsed[0].Translations) == 0 {
		return "", &ProviderError{Code: CodeUnknown, Detail: fmt.Sprintf("microsoft 响应无译文: %s", Truncate(string(raw), 200))}
	}
	return parsed[0].Translations[0].Text, nil
}

// getToken 取（或复用缓存的）Edge JWT。force 作废缓存强制重取。
func (p *MicrosoftProvider) getToken(ctx context.Context, force bool) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !force && p.token != "" && (!p.hasExpiry || time.Now().Add(edgeTokenLeeway).Before(p.tokenExp)) {
		return p.token, nil
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", p.authURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", NormalizeStreamError(err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", NormalizeStreamError(err)
	}
	if resp.StatusCode >= 400 {
		return "", &ProviderError{
			Code:       Classify(resp.StatusCode, string(raw)),
			StatusCode: resp.StatusCode,
			Detail:     "edge token 获取失败: " + ExtractErrorDetail(string(raw), 200),
		}
	}

	token := strings.TrimSpace(string(raw))
	if token == "" {
		return "", &ProviderError{Code: CodeUnknown, Detail: "edge token 为空"}
	}
	p.token = token
	p.tokenExp, p.hasExpiry = jwtExpiry(token)
	return token, nil
}

// jwtExpiry 解析 JWT payload 的 exp（秒）→ time；解析失败返回零值+false
// （无法判期的 token 按长缓存处理，下次 401 再强制刷新）。
func jwtExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Exp == 0 {
		return time.Time{}, false
	}
	return time.Unix(claims.Exp, 0), true
}
