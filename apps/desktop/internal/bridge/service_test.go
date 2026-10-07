package bridge

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"tako/internal/license"
	"tako/internal/provider"
	"tako/internal/store"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// fakeProvider 脚本化 provider（按 serviceID 之外的静态脚本跑）。
type fakeProvider struct {
	run func(emit func(*provider.StreamEvent) error) error
}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) TranslateStream(_ context.Context, _ provider.TextReq, emit func(*provider.StreamEvent) error) error {
	return f.run(emit)
}

// eventCollector 收集 (name, payload) 事件。
type eventCollector struct {
	mu     sync.Mutex
	events []*provider.StreamEvent
}

func (c *eventCollector) emit(name string, payload any) {
	if ev, ok := payload.(*provider.StreamEvent); ok {
		c.mu.Lock()
		c.events = append(c.events, ev)
		c.mu.Unlock()
	}
}

func (c *eventCollector) forService(serviceID string) []*provider.StreamEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	var out []*provider.StreamEvent
	for _, ev := range c.events {
		if ev.ServiceID == serviceID {
			out = append(out, ev)
		}
	}
	return out
}

// newTestService 构造注入内存 store + 收集器 emit 的 Service。
func newTestService(t *testing.T, factories map[string]providerFactory) (*Service, *store.Store, *eventCollector) {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "tako.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	col := &eventCollector{}
	s := &Service{
		store:     st,
		factories: factories,
		verifier:  license.FreeVerifier{},
		active:    map[string]context.CancelFunc{},
		pinned:    map[string]bool{},
		windows:   fakeWindows,
		emit:      col.emit,
	}
	return s, st, col
}

// fakeWindow 最小窗口句柄（仅记录 Hide 调用）。
type fakeWindow struct{ hidden bool }

func (f *fakeWindow) Hide() application.Window { f.hidden = true; return nil }

// fakeWindows 固定窗口表：input/popup 存在，其余 miss。
var fakeWindowRefs = map[string]*fakeWindow{
	"input": {},
	"popup": {},
}

func fakeWindows(name string) (windowRef, bool) {
	w, ok := fakeWindowRefs[name]
	if !ok {
		return nil, false
	}
	return w, true
}

// okFactory 返回总是成功的假 provider（start→delta→finish 序列）。
func okFactory(text string) providerFactory {
	return func(provider.ProviderConfig) (provider.Provider, error) {
		return &fakeProvider{run: func(emit func(*provider.StreamEvent) error) error {
			_ = emit(&provider.StreamEvent{Type: provider.EventMessageStart})
			_ = emit(&provider.StreamEvent{Type: provider.EventDelta, Text: text})
			return emit(&provider.StreamEvent{Type: provider.EventFinish, Text: text})
		}}, nil
	}
}

func failFactory(code provider.ErrorCode, detail string) providerFactory {
	return func(provider.ProviderConfig) (provider.Provider, error) {
		return &fakeProvider{run: func(emit func(*provider.StreamEvent) error) error {
			_ = emit(&provider.StreamEvent{Type: provider.EventMessageStart})
			return &provider.ProviderError{Code: code, Detail: detail}
		}}, nil
	}
}

func seedService(t *testing.T, st *store.Store, typ string, order int, enabled bool) store.Service {
	t.Helper()
	svc, err := st.SaveService(store.Service{
		Name: "svc-" + typ, Type: typ, Enabled: enabled,
		Config: store.ServiceConfig{Order: order},
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return svc
}

// waitEvents 等事件数达标（流式 fan-out 异步）。
func waitEvents(t *testing.T, col *eventCollector, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		col.mu.Lock()
		n := len(col.events)
		col.mu.Unlock()
		if n >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	col.mu.Lock()
	defer col.mu.Unlock()
	t.Fatalf("events=%d, want >= %d: %+v", len(col.events), want, col.events)
}

func TestTranslate_FanOut(t *testing.T) {
	factories := map[string]providerFactory{
		"llm":   okFactory("译文A"),
		"deepl": okFactory("译文B"),
	}
	s, st, col := newTestService(t, factories)
	svcA := seedService(t, st, "llm", 0, true)
	svcB := seedService(t, st, "deepl", 1, true)

	h, err := s.Translate(context.Background(), TextReq{Text: "hello", TargetLang: "zh-Hans"})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	if h.RequestID == "" {
		t.Fatal("empty requestId")
	}
	waitEvents(t, col, 6)

	for _, tc := range []struct {
		svc  store.Service
		want string
	}{{svcA, "译文A"}, {svcB, "译文B"}} {
		evs := col.forService(tc.svc.ID)
		if len(evs) != 3 {
			t.Fatalf("svc %s events=%d, want 3", tc.svc.ID, len(evs))
		}
		for _, ev := range evs {
			if ev.RequestID != h.RequestID || ev.ServiceID != tc.svc.ID {
				t.Fatalf("event ids wrong: %+v", ev)
			}
		}
		if evs[0].Type != provider.EventMessageStart ||
			evs[1].Type != provider.EventDelta || evs[1].Text != tc.want ||
			evs[2].Type != provider.EventFinish || evs[2].Text != tc.want {
			t.Fatalf("svc %s seq wrong: %+v", tc.svc.ID, evs)
		}
	}
}

func TestTranslate_DisabledAndMissing(t *testing.T) {
	s, st, col := newTestService(t, map[string]providerFactory{"llm": okFactory("x")})
	enabled := seedService(t, st, "llm", 0, true)
	disabled := seedService(t, st, "llm", 1, false)

	// 空 serviceIds → 只跑 enabled
	_, err := s.Translate(context.Background(), TextReq{Text: "hi", TargetLang: "en"})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	waitEvents(t, col, 3)
	if got := col.forService(disabled.ID); len(got) != 0 {
		t.Fatalf("disabled service 不应有事件: %+v", got)
	}
	if got := col.forService(enabled.ID); len(got) != 3 {
		t.Fatalf("enabled events=%d", len(got))
	}

	// 显式 serviceIds：不存在/禁用 → 各自终态 error
	col2 := &eventCollector{}
	s.emit = col2.emit
	h2, err := s.Translate(context.Background(), TextReq{
		Text: "hi", TargetLang: "en",
		ServiceIDs: []string{enabled.ID, disabled.ID, "nonexistent"},
	})
	if err != nil {
		t.Fatalf("Translate2: %v", err)
	}
	waitEvents(t, col2, 5) // enabled 3 + 2 errors
	for _, id := range []string{disabled.ID, "nonexistent"} {
		evs := col2.forService(id)
		if len(evs) != 1 || evs[0].Type != provider.EventError {
			t.Fatalf("svc %s 应只有终态 error: %+v", id, evs)
		}
		if evs[0].RequestID != h2.RequestID {
			t.Fatalf("event requestId wrong: %+v", evs[0])
		}
	}
}

func TestTranslate_EmptyText(t *testing.T) {
	s, _, _ := newTestService(t, nil)
	if _, err := s.Translate(context.Background(), TextReq{Text: "  "}); err == nil {
		t.Fatal("want error for empty text")
	}
}

func TestTranslate_ProviderFailureEmitsError(t *testing.T) {
	s, st, col := newTestService(t, map[string]providerFactory{
		"llm": failFactory(provider.CodeAuth, "invalid api key"),
	})
	svc := seedService(t, st, "llm", 0, true)

	if _, err := s.Translate(context.Background(), TextReq{Text: "hi", TargetLang: "en"}); err != nil {
		t.Fatalf("Translate: %v", err)
	}
	waitEvents(t, col, 2)
	evs := col.forService(svc.ID)
	last := evs[len(evs)-1]
	if last.Type != provider.EventError || last.ErrorCode != provider.CodeAuth {
		t.Fatalf("last=%+v, want error/auth", last)
	}
	if last.Message != "invalid api key" {
		t.Fatalf("message=%q", last.Message)
	}
}

func TestCancelTranslate(t *testing.T) {
	// provider 阻塞在 ctx.Done 上：取消后产出 error(cancelled)。
	factories := map[string]providerFactory{
		"llm": func(provider.ProviderConfig) (provider.Provider, error) {
			return &fakeProvider{run: func(emit func(*provider.StreamEvent) error) error {
				return errors.New("unreachable") // 不被走到（见下）
			}}, nil
		},
	}
	s, st, col := newTestService(t, factories)
	svc := seedService(t, st, "llm", 0, true)

	// 替换为挂起 provider：等 ctx 取消后返回 ctx.Err()
	s.factories["llm"] = func(provider.ProviderConfig) (provider.Provider, error) {
		return &fakeProvider{run: func(emit func(*provider.StreamEvent) error) error {
			return context.Canceled // StreamWithRetry 把 ctx.Err 上抛前 emit 无事件
		}}, nil
	}
	_ = svc

	h, err := s.Translate(context.Background(), TextReq{Text: "hi", TargetLang: "en"})
	if err != nil {
		t.Fatalf("Translate: %v", err)
	}
	waitEvents(t, col, 1) // error(cancelled)
	evs := col.events
	if evs[0].Type != provider.EventError || evs[0].ErrorCode != provider.CodeCancelled {
		t.Fatalf("want cancelled error, got %+v", evs[0])
	}
	if err := s.CancelTranslate(h.RequestID); err != nil {
		t.Fatalf("CancelTranslate: %v", err)
	}
	// 已结束的 requestID 再取消——幂等 nil
	if err := s.CancelTranslate(h.RequestID); err != nil {
		t.Fatalf("re-cancel: %v", err)
	}
}

func TestTestService(t *testing.T) {
	s, _, _ := newTestService(t, map[string]providerFactory{
		"llm": okFactory("你好"),
	})
	res, err := s.TestService(context.Background(), ServiceSetting{Type: "llm"})
	if err != nil || !res.OK || res.Sample != "你好" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if res.LatencyMs < 0 {
		t.Fatal("latency negative")
	}

	res2, _ := s.TestService(context.Background(), ServiceSetting{Type: "nope"})
	if res2.OK || res2.Error == "" {
		t.Fatalf("unknown type res=%+v", res2)
	}
}

func TestServicesAndHistoryPassthrough(t *testing.T) {
	s, st, _ := newTestService(t, nil)

	saved, err := s.SaveService(ServiceSetting{Name: "X", Type: "deepl", Enabled: true, Order: 3})
	if err != nil || saved.ID == "" {
		t.Fatalf("save: %+v %v", saved, err)
	}
	list, err := s.ListServices()
	if err != nil || len(list) != 1 || list[0].ID != saved.ID {
		t.Fatalf("list: %+v %v", list, err)
	}
	if err := s.DeleteService(saved.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if list, _ := s.ListServices(); len(list) != 0 {
		t.Fatal("after delete")
	}
	_ = st

	h, err := s.SaveHistory(HistoryItem{SourceText: "a", TargetLang: "zh-Hans", ServiceID: "s1", ResultText: "b"})
	if err != nil || h.ID == "" {
		t.Fatalf("saveHistory: %+v %v", h, err)
	}
	page, err := s.ListHistory(HistoryQuery{})
	if err != nil || page.Total != 1 {
		t.Fatalf("listHistory: %+v %v", page, err)
	}
	if err := s.ToggleFavorite(h.ID, true); err != nil {
		t.Fatalf("fav: %v", err)
	}
	page, _ = s.ListHistory(HistoryQuery{FavoriteOnly: true})
	if page.Total != 1 || !page.Items[0].Favorite {
		t.Fatalf("fav query: %+v", page)
	}
	if err := s.DeleteHistory(h.ID); err != nil {
		t.Fatalf("deleteHistory: %v", err)
	}
}

func TestSettingsRoundtrip(t *testing.T) {
	s, _, _ := newTestService(t, nil)

	def, err := s.GetSettings()
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if def.TargetLang != "zh-Hans" || def.MaxConcurrent != 5 || def.Hotkeys.SelectionTranslate != "Alt+D" {
		t.Fatalf("defaults: %+v", def)
	}

	settings := def
	settings.Theme = "dark"
	settings.MaxConcurrent = 99 // 超界应被钳到 10
	if err := s.SaveSettings(settings); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := s.GetSettings()
	if err != nil {
		t.Fatalf("get2: %v", err)
	}
	if got.Theme != "dark" || got.MaxConcurrent != 10 {
		t.Fatalf("roundtrip: %+v", got)
	}
}

func TestLicenseInfo(t *testing.T) {
	s, _, _ := newTestService(t, nil)
	li := s.LicenseInfo()
	if !li.Licensed || li.Plan != "free" {
		t.Fatalf("license=%+v", li)
	}
}

func TestWindowPinned(t *testing.T) {
	s, _, _ := newTestService(t, nil)
	if s.IsPinned("input") {
		t.Fatal("默认不应钉住")
	}
	if err := s.SetWindowPinned("input", true); err != nil {
		t.Fatalf("pin: %v", err)
	}
	if !s.IsPinned("input") {
		t.Fatal("pin 后 IsPinned 应为 true")
	}
	if err := s.SetWindowPinned("input", false); err != nil {
		t.Fatalf("unpin: %v", err)
	}
	if s.IsPinned("input") {
		t.Fatal("unpin 后 IsPinned 应为 false")
	}
	// 未登记窗口名报错
	if err := s.SetWindowPinned("nope", true); err == nil {
		t.Fatal("未知窗口应报错")
	}
}

func TestHideWindow(t *testing.T) {
	s, _, _ := newTestService(t, nil)
	fakeWindowRefs["popup"].hidden = false
	if err := s.HideWindow("popup"); err != nil {
		t.Fatalf("hide: %v", err)
	}
	if !fakeWindowRefs["popup"].hidden {
		t.Fatal("窗口应已隐藏")
	}
	if err := s.HideWindow("nope"); err == nil {
		t.Fatal("未知窗口应报错")
	}
}
