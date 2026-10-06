package llm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"tako/internal/provider"
)

// collectEmit 收集 emit 的事件序列。
type collectEmit struct{ events []*provider.StreamEvent }

func (c *collectEmit) emit(ev *provider.StreamEvent) error {
	c.events = append(c.events, ev)
	return nil
}

func newTestProvider(t *testing.T, server *httptest.Server) *OpenAIProvider {
	t.Helper()
	p, err := New(provider.ProviderConfig{
		Type:    "llm",
		BaseURL: server.URL + "/v1",
		APIKey:  "sk-test",
		Model:   "test-model",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return p
}

func sseHandler(chunks ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			http.Error(w, "bad path", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer sk-test" {
			http.Error(w, `{"error":{"message":"invalid api key"}}`, http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, _ := w.(http.Flusher)
		for _, c := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", c)
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

func deltaChunk(text string) string {
	return fmt.Sprintf(`{"id":"c1","choices":[{"delta":{"content":%q}}]}`, text)
}

// 正常流：start → delta* → finish（全量 text + usage）。
func TestOpenAIProvider_NormalStream(t *testing.T) {
	srv := httptest.NewServer(sseHandler(
		deltaChunk("Hello"),
		deltaChunk(" world"),
		`{"id":"c1","choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":10,"completion_tokens":5}}`,
		`[DONE]`,
	))
	defer srv.Close()

	c := &collectEmit{}
	err := newTestProvider(t, srv).TranslateStream(context.Background(),
		provider.TextReq{Text: "你好", TargetLang: "en"}, c.emit)
	if err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	if len(c.events) != 4 {
		t.Fatalf("events=%d, want 4: %+v", len(c.events), c.events)
	}
	if c.events[0].Type != provider.EventMessageStart {
		t.Fatalf("events[0]=%s, want message-start", c.events[0].Type)
	}
	if c.events[1].Type != provider.EventDelta || c.events[1].Text != "Hello" {
		t.Fatalf("events[1]=%+v", c.events[1])
	}
	fin := c.events[3]
	if fin.Type != provider.EventFinish || fin.Text != "Hello world" {
		t.Fatalf("finish=%+v", fin)
	}
	if fin.Usage == nil || fin.Usage.PromptTokens != 10 || fin.Usage.CompletionTokens != 5 {
		t.Fatalf("usage=%+v", fin.Usage)
	}
}

// 流内 error 字段（任意位置）→ 归类上抛。
func TestOpenAIProvider_StreamErrorChunk(t *testing.T) {
	srv := httptest.NewServer(sseHandler(
		deltaChunk("Hel"),
		`{"error":{"message":"too many requests","type":"rate_limit_error"}}`,
	))
	defer srv.Close()

	c := &collectEmit{}
	err := newTestProvider(t, srv).TranslateStream(context.Background(),
		provider.TextReq{Text: "x", TargetLang: "en"}, c.emit)
	if err == nil {
		t.Fatal("want error")
	}
	if code := provider.ErrorCodeOf(err); code != provider.CodeRateLimit {
		t.Fatalf("code=%s, want rate_limit", code)
	}
}

// 干净 EOF 但无终态标记 → connect 断流错误；已产出 delta 不得合成 finish。
func TestOpenAIProvider_TruncatedStream(t *testing.T) {
	srv := httptest.NewServer(sseHandler(deltaChunk("partial")))
	defer srv.Close()

	c := &collectEmit{}
	err := newTestProvider(t, srv).TranslateStream(context.Background(),
		provider.TextReq{Text: "x", TargetLang: "en"}, c.emit)
	if err == nil {
		t.Fatal("want error")
	}
	if code := provider.ErrorCodeOf(err); code != provider.CodeConnect {
		t.Fatalf("code=%s, want connect", code)
	}
	for _, ev := range c.events {
		if ev.Type == provider.EventFinish {
			t.Fatal("truncated stream must not emit finish")
		}
	}
}

// 401 → auth，且不判为瞬时错误。
func TestOpenAIProvider_Unauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid api key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := &collectEmit{}
	err := newTestProvider(t, srv).TranslateStream(context.Background(),
		provider.TextReq{Text: "x", TargetLang: "en"}, c.emit)
	if provider.ErrorCodeOf(err) != provider.CodeAuth {
		t.Fatalf("code=%s, want auth", provider.ErrorCodeOf(err))
	}
	if provider.IsTransient(err) {
		t.Fatal("auth 不应判为瞬时错误")
	}
}

// 鉴权头：customHeaders 不得伪造 authorization/x-api-key，由 apiKey 重建。
func TestOpenAIProvider_AuthHeaderRebuild(t *testing.T) {
	var gotAuth, gotKey, gotCustom string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotKey = r.Header.Get("x-api-key")
		gotCustom = r.Header.Get("X-Team")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	p, err := New(provider.ProviderConfig{
		Type:    "llm",
		BaseURL: srv.URL + "/v1",
		APIKey:  "real-key",
		Model:   "m",
		CustomHeaders: map[string]string{
			"X-Team":        "blue",
			"Authorization": "Bearer forged",
			"X-Api-Key":     "forged",
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c := &collectEmit{}
	if err := p.TranslateStream(context.Background(),
		provider.TextReq{Text: "x", TargetLang: "en"}, c.emit); err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	if gotAuth != "Bearer real-key" {
		t.Fatalf("Authorization=%q, want Bearer real-key", gotAuth)
	}
	if gotKey != "" {
		t.Fatalf("x-api-key=%q, want empty (forged stripped)", gotKey)
	}
	if gotCustom != "blue" {
		t.Fatalf("X-Team=%q", gotCustom)
	}
}

// x-api-key 鉴权方案。
func TestOpenAIProvider_XAPIKeyScheme(t *testing.T) {
	var gotAuth, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotKey = r.Header.Get("x-api-key")
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	p, err := New(provider.ProviderConfig{
		Type:       "llm",
		BaseURL:    srv.URL + "/v1",
		APIKey:     "ak-1",
		Model:      "m",
		AuthScheme: "x-api-key",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	c := &collectEmit{}
	if err := p.TranslateStream(context.Background(),
		provider.TextReq{Text: "x", TargetLang: "en"}, c.emit); err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	if gotKey != "ak-1" || gotAuth != "" {
		t.Fatalf("x-api-key=%q Authorization=%q", gotKey, gotAuth)
	}
}
