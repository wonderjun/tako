package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// assertOnceEvents 校验传统服务一次性事件序列：start → delta{text} → finish{text}。
func assertOnceEvents(t *testing.T, events []*StreamEvent, wantText string) {
	t.Helper()
	if len(events) != 3 {
		t.Fatalf("events=%d, want 3: %+v", len(events), events)
	}
	if events[0].Type != EventMessageStart {
		t.Fatalf("events[0]=%s", events[0].Type)
	}
	if events[1].Type != EventDelta || events[1].Text != wantText {
		t.Fatalf("delta=%+v", events[1])
	}
	if events[2].Type != EventFinish || events[2].Text != wantText {
		t.Fatalf("finish=%+v", events[2])
	}
}

// ---------------------------------------------------------------------------
// DeepL
// ---------------------------------------------------------------------------

func TestDeepL_RequestShapeAndParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/translate" {
			t.Errorf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "DeepL-Auth-Key dk-1" {
			t.Errorf("Authorization=%q", got)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		if r.Form.Get("target_lang") != "ZH" {
			t.Errorf("target_lang=%q, want ZH", r.Form.Get("target_lang"))
		}
		if r.Form.Get("source_lang") != "EN" {
			t.Errorf("source_lang=%q, want EN", r.Form.Get("source_lang"))
		}
		if r.Form.Get("text") != "hello" {
			t.Errorf("text=%q", r.Form.Get("text"))
		}
		w.Write([]byte(`{"translations":[{"detected_source_language":"EN","text":"你好"}]}`))
	}))
	defer srv.Close()

	p, err := NewDeepL(ProviderConfig{Type: "deepl", APIKey: "dk-1", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewDeepL: %v", err)
	}
	c := &collectEmit{}
	err = p.TranslateStream(context.Background(),
		TextReq{Text: "hello", SourceLang: "en", TargetLang: "zh-Hans"}, c.emit)
	if err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	assertOnceEvents(t, c.events, "你好")
}

func TestDeepL_QuotaExceeded456(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Quota exceeded"}`, 456)
	}))
	defer srv.Close()

	p, _ := NewDeepL(ProviderConfig{Type: "deepl", APIKey: "k", BaseURL: srv.URL})
	c := &collectEmit{}
	err := p.TranslateStream(context.Background(),
		TextReq{Text: "x", SourceLang: "auto", TargetLang: "en"}, c.emit)
	if ErrorCodeOf(err) != CodeRateLimit {
		t.Fatalf("code=%s, want rate_limit", ErrorCodeOf(err))
	}
}

func TestDeepL_MissingKey(t *testing.T) {
	_, err := NewDeepL(ProviderConfig{Type: "deepl"})
	if ErrorCodeOf(err) != CodeAuth {
		t.Fatalf("code=%s, want auth", ErrorCodeOf(err))
	}
}

// ---------------------------------------------------------------------------
// Google
// ---------------------------------------------------------------------------

func TestGoogle_RequestShapeAndParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/translate_a/single" {
			t.Errorf("path=%s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("client") != "gtx" || q.Get("dt") != "t" {
			t.Errorf("query=%s", r.URL.RawQuery)
		}
		if q.Get("sl") != "auto" || q.Get("tl") != "zh-CN" {
			t.Errorf("sl=%q tl=%q, want auto/zh-CN", q.Get("sl"), q.Get("tl"))
		}
		// 嵌套数组：[[["译文1","原文1",...],["译文2","原文2",...]],null,"en",...]
		w.Write([]byte(`[[["你好","hello",null,null,10],["世界","world",null,null,1]],null,"en"]`))
	}))
	defer srv.Close()

	p, err := NewGoogle(ProviderConfig{Type: "google", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewGoogle: %v", err)
	}
	c := &collectEmit{}
	err = p.TranslateStream(context.Background(),
		TextReq{Text: "hello world", SourceLang: "auto", TargetLang: "zh-Hans"}, c.emit)
	if err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	assertOnceEvents(t, c.events, "你好世界")
}

// ---------------------------------------------------------------------------
// Microsoft（Edge 免 key）
// ---------------------------------------------------------------------------

// fakeJWT 构造测试用 JWT（payload 带 exp）。
func fakeJWT(exp time.Time) string {
	payload, _ := json.Marshal(map[string]any{"exp": exp.Unix()})
	return base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`)) + "." +
		base64.RawURLEncoding.EncodeToString(payload) + ".sig"
}

func TestMicrosoft_TokenFlowAndParse(t *testing.T) {
	var authCalls atomic.Int32
	token := fakeJWT(time.Now().Add(10 * time.Minute))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/translate/auth":
			authCalls.Add(1)
			w.Write([]byte(token))
		case "/translate":
			if got := r.Header.Get("Authorization"); got != "Bearer "+token {
				t.Errorf("Authorization=%q", got)
			}
			if r.URL.Query().Get("to") != "zh-Hans" || r.URL.Query().Get("from") != "en" {
				t.Errorf("query=%s", r.URL.RawQuery)
			}
			w.Write([]byte(`[{"translations":[{"text":"你好","to":"zh-Hans"}]}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	p, err := NewMicrosoft(ProviderConfig{Type: "microsoft", BaseURL: srv.URL})
	if err != nil {
		t.Fatalf("NewMicrosoft: %v", err)
	}
	p.authURL = srv.URL + "/translate/auth" // 测试覆盖 auth 端点

	c := &collectEmit{}
	err = p.TranslateStream(context.Background(),
		TextReq{Text: "hello", SourceLang: "en", TargetLang: "zh-Hans"}, c.emit)
	if err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	assertOnceEvents(t, c.events, "你好")
	if authCalls.Load() != 1 {
		t.Fatalf("authCalls=%d, want 1", authCalls.Load())
	}

	// 第二次调用应复用缓存 token（不再请求 auth）
	c2 := &collectEmit{}
	if err := p.TranslateStream(context.Background(),
		TextReq{Text: "again", SourceLang: "en", TargetLang: "zh-Hans"}, c2.emit); err != nil {
		t.Fatalf("second: %v", err)
	}
	if authCalls.Load() != 1 {
		t.Fatalf("authCalls=%d, want 1（token 缓存命中）", authCalls.Load())
	}
}

// 401 → 作废 token 重取一次再成功。
func TestMicrosoft_TokenRefreshOn401(t *testing.T) {
	var authCalls atomic.Int32
	var translateCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/translate/auth":
			authCalls.Add(1)
			w.Write([]byte(fakeJWT(time.Now().Add(10 * time.Minute))))
		case "/translate":
			if translateCalls.Add(1) == 1 {
				http.Error(w, `{"error":{"code":401000,"message":"token expired"}}`, http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`[{"translations":[{"text":"OK","to":"en"}]}]`))
		}
	}))
	defer srv.Close()

	p, _ := NewMicrosoft(ProviderConfig{Type: "microsoft", BaseURL: srv.URL})
	p.authURL = srv.URL + "/translate/auth"

	c := &collectEmit{}
	err := p.TranslateStream(context.Background(),
		TextReq{Text: "x", SourceLang: "zh-Hans", TargetLang: "en"}, c.emit)
	if err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
	assertOnceEvents(t, c.events, "OK")
	if authCalls.Load() != 2 || translateCalls.Load() != 2 {
		t.Fatalf("auth=%d translate=%d, want 2/2", authCalls.Load(), translateCalls.Load())
	}
}

// auto 源语言不携带 from 参数。
func TestMicrosoft_AutoSourceOmitsFrom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/translate/auth":
			w.Write([]byte(fakeJWT(time.Now().Add(10 * time.Minute))))
		case "/translate":
			if r.URL.Query().Has("from") {
				t.Errorf("auto 源不应带 from: %s", r.URL.RawQuery)
			}
			w.Write([]byte(`[{"translations":[{"text":"x","to":"zh-Hans"}]}]`))
		}
	}))
	defer srv.Close()

	p, _ := NewMicrosoft(ProviderConfig{Type: "microsoft", BaseURL: srv.URL})
	p.authURL = srv.URL + "/translate/auth"
	c := &collectEmit{}
	if err := p.TranslateStream(context.Background(),
		TextReq{Text: "x", SourceLang: "auto", TargetLang: "zh-Hans"}, c.emit); err != nil {
		t.Fatalf("TranslateStream: %v", err)
	}
}
