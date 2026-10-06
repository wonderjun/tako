// BridgeService 是 Wails 暴露给前端的唯一服务：方法名与参数结构对齐
// packages/contracts 的 BridgeMethods。
//
// 装配模型（main.go 零改动）：NewService() 无参恒定，依赖在
// ServiceStartup（Wails 生命周期钩子）里装配；ServiceShutdown 释放。
// 测试经同包字段注入（store/factories/emit/verifier）。
package bridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/wailsapp/wails/v3/pkg/application"

	"tako/internal/license"
	"tako/internal/provider"
	"tako/internal/selection"
	"tako/internal/store"
)

// translateMaxRetries 单服务子流的自动重连预算（翻译场景 2 次足够，
// 重试事件 attempt 从 1 递增）。
const translateMaxRetries = 2

// testServiceTimeout 连通性测试的超时上限。
const testServiceTimeout = 20 * time.Second

// Service bridge 服务实例。
type Service struct {
	mu        sync.Mutex
	store     *store.Store
	factories map[string]providerFactory
	verifier  license.Verifier
	active    map[string]context.CancelFunc // requestID → cancel
	emit      emitFunc
}

// NewService 构造 bridge 服务（无参恒定：main.go 装配行永远不变，
// 真实依赖在 ServiceStartup 注入默认值）。
func NewService() *Service {
	return &Service{
		factories: defaultFactories(),
		verifier:  license.FreeVerifier{},
		active:    map[string]context.CancelFunc{},
		emit:      wailsEmit,
	}
}

// wailsEmit 生产事件出口（wails 应用单例）。
func wailsEmit(name string, payload any) {
	application.Get().Event.Emit(name, payload)
}

// ServiceStartup Wails 启动服务回调：打开默认库完成装配；
// 装配失败返回 error 使应用启动即败露。
func (s *Service) ServiceStartup(context.Context, application.ServiceOptions) error {
	st, err := store.OpenDefault()
	if err != nil {
		return err
	}
	if err := seedDefaultServices(st); err != nil {
		_ = st.Close()
		return err
	}
	s.mu.Lock()
	s.store = st
	s.mu.Unlock()
	return nil
}

// ServiceShutdown Wails 关停回调：取消全部在途翻译并关闭库。
func (s *Service) ServiceShutdown() {
	s.mu.Lock()
	for _, cancel := range s.active {
		cancel()
	}
	st := s.store
	s.mu.Unlock()
	if st != nil {
		_ = st.Close()
	}
}

// Ping 链路验证，返回 pong 标识。
func (s *Service) Ping() string { return "pong:tako" }

var errNotReady = errors.New("bridge: 服务未初始化（ServiceStartup 未完成）")

func (s *Service) requireStore() (*store.Store, error) {
	s.mu.Lock()
	st := s.store
	s.mu.Unlock()
	if st == nil {
		return nil, errNotReady
	}
	return st, nil
}

// ---------------------------------------------------------------------------
// 翻译（fan-out 多服务子流）
// ---------------------------------------------------------------------------

// Translate 发起多服务翻译，事件经 translate:event 推送。
// serviceIds 为空时取全部 enabled 服务（按 order 升序）。
func (s *Service) Translate(ctx context.Context, req TextReq) (TranslateHandle, error) {
	st, err := s.requireStore()
	if err != nil {
		return TranslateHandle{}, err
	}
	if strings.TrimSpace(req.Text) == "" {
		return TranslateHandle{}, errors.New("bridge: text 为空")
	}

	services, missing, err := s.resolveServices(st, req.ServiceIDs)
	if err != nil {
		return TranslateHandle{}, err
	}

	requestID := newID()
	// RPC 立即返回而流继续：runCtx 必须脱离请求级 ctx 的取消
	// （WithoutCancel 保留值、丢弃取消），由 active map 统一管理。
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	s.mu.Lock()
	s.active[requestID] = cancel
	s.mu.Unlock()

	var wg sync.WaitGroup
	// 请求的 serviceId 不存在/已禁用：该子流直接终态 error。
	for _, id := range missing {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			s.emitError(requestID, id, provider.CodeUnknown, "服务不存在或已禁用")
		}(id)
	}
	textReq := provider.TextReq{
		Text:       req.Text,
		SourceLang: req.SourceLang,
		TargetLang: req.TargetLang,
	}
	for _, svc := range services {
		wg.Add(1)
		go func(svc store.Service) {
			defer wg.Done()
			s.runServiceStream(runCtx, requestID, svc, textReq)
		}(svc)
	}

	go func() {
		wg.Wait()
		s.mu.Lock()
		delete(s.active, requestID)
		s.mu.Unlock()
	}()
	return TranslateHandle{RequestID: requestID}, nil
}

// resolveServices 解析本次参与翻译的服务：空 ids → 全部 enabled；
// 显式 ids → 逐个查库，不存在/禁用的归入 missing。
func (s *Service) resolveServices(st *store.Store, ids []string) ([]store.Service, []string, error) {
	if len(ids) == 0 {
		all, err := st.ListServices()
		if err != nil {
			return nil, nil, err
		}
		var out []store.Service
		for _, svc := range all {
			if svc.Enabled {
				out = append(out, svc)
			}
		}
		return out, nil, nil
	}
	var found []store.Service
	var missing []string
	for _, id := range ids {
		svc, err := st.GetService(id)
		if errors.Is(err, store.ErrNotFound) {
			missing = append(missing, id)
			continue
		}
		if err != nil {
			return nil, nil, err
		}
		if !svc.Enabled {
			missing = append(missing, id)
			continue
		}
		found = append(found, svc)
	}
	return found, missing, nil
}

// runServiceStream 单服务子流：构造 provider → StreamWithRetry →
// 终态失败时 emit error 事件（成功路径的 finish 由 provider 发出）。
func (s *Service) runServiceStream(ctx context.Context, requestID string, svc store.Service, req provider.TextReq) {
	emit := func(ev *provider.StreamEvent) error {
		ev.RequestID = requestID
		ev.ServiceID = svc.ID
		s.emit(eventTranslateEvent, ev)
		return nil
	}
	p, err := s.buildProvider(svc)
	if err != nil {
		emitErrorEvent(emit, err)
		return
	}
	if err := provider.StreamWithRetry(ctx, p, req, emit, translateMaxRetries); err != nil {
		emitErrorEvent(emit, err)
	}
}

// emitErrorEvent 发终态 error 事件（errorCode + 用户可读 message）。
func emitErrorEvent(emit func(*provider.StreamEvent) error, err error) {
	_ = emit(&provider.StreamEvent{
		Type:      provider.EventError,
		ErrorCode: provider.ErrorCodeOf(err),
		Message:   providerErrorMessage(err),
	})
}

// emitError 针对指定 serviceId 直接发终态 error 事件。
func (s *Service) emitError(requestID, serviceID string, code provider.ErrorCode, message string) {
	s.emit(eventTranslateEvent, &provider.StreamEvent{
		RequestID: requestID,
		ServiceID: serviceID,
		Type:      provider.EventError,
		ErrorCode: code,
		Message:   message,
	})
}

// providerErrorMessage 提取用户可读错误文案：ProviderError 用 Detail
// （已脱敏），其余取错误原文，截断 300 字符。
func providerErrorMessage(err error) string {
	var pe *provider.ProviderError
	if errors.As(err, &pe) && pe.Detail != "" {
		return provider.Truncate(pe.Detail, 300)
	}
	return provider.Truncate(err.Error(), 300)
}

// CancelTranslate 取消一次进行中的翻译（幂等：已结束的 requestID 不算错）。
// 取消后各子流经 ctx.Done 路径发出 error(errorCode=cancelled) 终态事件。
func (s *Service) CancelTranslate(requestID string) error {
	s.mu.Lock()
	cancel := s.active[requestID]
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

// ---------------------------------------------------------------------------
// 连通性测试
// ---------------------------------------------------------------------------

// TestService 用给定配置（未保存也可）做一次最小翻译验证连通性。
func (s *Service) TestService(ctx context.Context, service ServiceSetting) (TestResult, error) {
	factory, ok := s.factories[service.Type]
	if !ok {
		return TestResult{OK: false, Error: "未知服务类型: " + service.Type}, nil
	}
	p, err := factory(settingToConfig(service))
	if err != nil {
		return TestResult{OK: false, Error: providerErrorMessage(err)}, nil
	}
	target := "zh-Hans"
	if st, err := s.requireStore(); err == nil {
		target = s.currentTargetLang(st)
	}

	tctx, cancel := context.WithTimeout(ctx, testServiceTimeout)
	defer cancel()
	start := time.Now()
	var full strings.Builder
	err = p.TranslateStream(tctx, provider.TextReq{
		Text: "hello", SourceLang: "auto", TargetLang: target,
	}, func(ev *provider.StreamEvent) error {
		if ev.Type == provider.EventFinish {
			full.WriteString(ev.Text)
		}
		return nil
	})
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return TestResult{OK: false, LatencyMs: latency, Error: providerErrorMessage(err)}, nil
	}
	return TestResult{OK: true, LatencyMs: latency, Sample: full.String()}, nil
}

// currentTargetLang 读设置里的目标语言（缺省 zh-Hans）。
func (s *Service) currentTargetLang(st *store.Store) string {
	settings, err := s.getSettings(st)
	if err != nil || settings.TargetLang == "" {
		return "zh-Hans"
	}
	return settings.TargetLang
}

// ---------------------------------------------------------------------------
// 服务配置 CRUD
// ---------------------------------------------------------------------------

func (s *Service) ListServices() ([]ServiceSetting, error) {
	st, err := s.requireStore()
	if err != nil {
		return nil, err
	}
	rows, err := st.ListServices()
	if err != nil {
		return nil, err
	}
	out := make([]ServiceSetting, 0, len(rows))
	for _, r := range rows {
		out = append(out, fromStoreService(r))
	}
	return out, nil
}

func (s *Service) SaveService(service ServiceSetting) (ServiceSetting, error) {
	st, err := s.requireStore()
	if err != nil {
		return ServiceSetting{}, err
	}
	saved, err := st.SaveService(toStoreService(service))
	if err != nil {
		return ServiceSetting{}, err
	}
	return fromStoreService(saved), nil
}

func (s *Service) DeleteService(id string) error {
	st, err := s.requireStore()
	if err != nil {
		return err
	}
	return st.DeleteService(id)
}

// ---------------------------------------------------------------------------
// 历史
// ---------------------------------------------------------------------------

func (s *Service) ListHistory(query HistoryQuery) (HistoryPage, error) {
	st, err := s.requireStore()
	if err != nil {
		return HistoryPage{}, err
	}
	items, total, err := st.ListHistory(store.HistoryQuery{
		Keyword:      query.Keyword,
		ServiceID:    query.ServiceID,
		FavoriteOnly: query.FavoriteOnly,
		Limit:        query.Limit,
		Offset:       query.Offset,
	})
	if err != nil {
		return HistoryPage{}, err
	}
	out := HistoryPage{Items: make([]HistoryItem, 0, len(items)), Total: total}
	for _, it := range items {
		out.Items = append(out.Items, HistoryItem{
			ID:          it.ID,
			CreatedAt:   it.CreatedAt,
			SourceText:  it.SourceText,
			SourceLang:  it.SourceLang,
			TargetLang:  it.TargetLang,
			ServiceID:   it.ServiceID,
			ServiceName: it.ServiceName,
			ResultText:  it.ResultText,
			Favorite:    it.Favorite,
		})
	}
	return out, nil
}

func (s *Service) SaveHistory(item HistoryItem) (HistoryItem, error) {
	st, err := s.requireStore()
	if err != nil {
		return HistoryItem{}, err
	}
	saved, err := st.SaveHistory(store.HistoryItem{
		ID:          item.ID,
		CreatedAt:   item.CreatedAt,
		SourceText:  item.SourceText,
		SourceLang:  item.SourceLang,
		TargetLang:  item.TargetLang,
		ServiceID:   item.ServiceID,
		ServiceName: item.ServiceName,
		ResultText:  item.ResultText,
		Favorite:    item.Favorite,
	})
	if err != nil {
		return HistoryItem{}, err
	}
	item.ID = saved.ID
	item.CreatedAt = saved.CreatedAt
	return item, nil
}

func (s *Service) DeleteHistory(id string) error {
	st, err := s.requireStore()
	if err != nil {
		return err
	}
	return st.DeleteHistory(id)
}

func (s *Service) ToggleFavorite(id string, favorite bool) error {
	st, err := s.requireStore()
	if err != nil {
		return err
	}
	return st.ToggleFavorite(id, favorite)
}

func (s *Service) ClearHistory() error {
	st, err := s.requireStore()
	if err != nil {
		return err
	}
	return st.ClearHistory()
}

// ---------------------------------------------------------------------------
// 设置 / 权限 / 许可证
// ---------------------------------------------------------------------------

const settingsKeyApp = "app"

// settingsKeySeeded 首启播种标记：免 key 服务仅在首次空库时播种一次，
// 用户此后删光服务不会被复活打扰。
const settingsKeySeeded = "servicesSeeded"

// seedDefaultServices 首启播种免 key 服务（google/microsoft 均无需 apiKey），
// 让首次运行即得真实翻译与双渠道对照；已播过或服务表非空时跳过。
func seedDefaultServices(st *store.Store) error {
	_, seeded, err := st.GetSetting(settingsKeySeeded)
	if err != nil {
		return err
	}
	if seeded {
		return nil
	}
	services, err := st.ListServices()
	if err != nil {
		return err
	}
	if len(services) == 0 {
		seeds := []store.Service{
			{Name: "Google 翻译", Type: "google", Enabled: true, Config: store.ServiceConfig{Order: 0}},
			{Name: "Microsoft 翻译", Type: "microsoft", Enabled: true, Config: store.ServiceConfig{Order: 1}},
		}
		for _, svc := range seeds {
			if _, err := st.SaveService(svc); err != nil {
				return err
			}
		}
	}
	// 表非空（用户已有配置）同样置位：播种只弥补"首次零服务"的开箱空窗。
	return st.SetSetting(settingsKeySeeded, json.RawMessage(`true`))
}

// defaultAppSettings 缺省设置（与前端 mock 一致）。
func defaultAppSettings() AppSettings {
	return AppSettings{
		SourceLang:    "auto",
		TargetLang:    "zh-Hans",
		Hotkeys:       HotkeyConfig{SelectionTranslate: "Alt+D", InputTranslate: "Alt+A"},
		Theme:         "system",
		MaxConcurrent: 5,
	}
}

// getSettings 读设置（无记录返回默认值）。
func (s *Service) getSettings(st *store.Store) (AppSettings, error) {
	settings := defaultAppSettings()
	raw, found, err := st.GetSetting(settingsKeyApp)
	if err != nil || !found {
		return settings, err
	}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return defaultAppSettings(), nil // 脏数据回退默认，不阻断启动
	}
	return settings, nil
}

func (s *Service) GetSettings() (AppSettings, error) {
	st, err := s.requireStore()
	if err != nil {
		return AppSettings{}, err
	}
	return s.getSettings(st)
}

func (s *Service) SaveSettings(settings AppSettings) error {
	st, err := s.requireStore()
	if err != nil {
		return err
	}
	// 钳制多开上限（契约：1-10）。
	if settings.MaxConcurrent < 1 {
		settings.MaxConcurrent = 1
	}
	if settings.MaxConcurrent > 10 {
		settings.MaxConcurrent = 10
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	return st.SetSetting(settingsKeyApp, raw)
}

// CheckAccessibility macOS 辅助功能权限（模拟按键取词依赖）；Windows 恒 true。
func (s *Service) CheckAccessibility() bool { return selection.AccessibilityTrusted(false) }

// OpenAccessibilitySettings 打开系统辅助功能设置面板。
func (s *Service) OpenAccessibilitySettings() error { return selection.OpenAccessibilitySettings() }

// LicenseInfo 许可证状态（一期 FreeVerifier 恒 free）。
func (s *Service) LicenseInfo() LicenseInfo {
	info := s.verifier.Verify()
	out := LicenseInfo{Licensed: info.Licensed, Plan: info.Plan}
	if !info.ExpiresAt.IsZero() {
		out.ExpiresAt = info.ExpiresAt.UnixMilli()
	}
	return out
}

// ---------------------------------------------------------------------------
// store ↔ bridge 映射与工具
// ---------------------------------------------------------------------------

func toStoreService(s ServiceSetting) store.Service {
	return store.Service{
		ID:        s.ID,
		Name:      s.Name,
		Type:      s.Type,
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		Config: store.ServiceConfig{
			BaseURL:       s.BaseURL,
			APIKey:        s.APIKey,
			Model:         s.Model,
			AuthScheme:    s.AuthScheme,
			CustomHeaders: s.CustomHeaders,
			Order:         s.Order,
		},
	}
}

func fromStoreService(s store.Service) ServiceSetting {
	return ServiceSetting{
		ID:            s.ID,
		Name:          s.Name,
		Type:          s.Type,
		Enabled:       s.Enabled,
		BaseURL:       s.Config.BaseURL,
		APIKey:        s.Config.APIKey,
		Model:         s.Config.Model,
		AuthScheme:    s.Config.AuthScheme,
		CustomHeaders: s.Config.CustomHeaders,
		Order:         s.Config.Order,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}
}

// newID 生成请求 ID（nanoid）。
func newID() string {
	id, err := gonanoid.New()
	if err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return "req-" + id
}
