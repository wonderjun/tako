// Service 是 Wails 暴露给前端的唯一服务：方法名与参数结构对齐
// packages/contracts 的 BridgeMethods（P0 全 stub，A5 实装）。
//
// 依赖装配约定（A5，见 docs/spec/SPEC-A.md）：
// NewService() 保持无参（main.go 主线独占、永不因后端改动而变）；
// store / provider 注册表 / active map 全部在 ServiceStartup 钩子内装配，
// ServiceShutdown 释放。事件推送用 application.Get().Event.Emit。
package bridge

import (
	"context"
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Service bridge 服务实例。依赖（store、provider 注册表、active map）
// 在 ServiceStartup 内装配，P0 保持零依赖。
type Service struct{}

// NewService 构造 bridge 服务（无参：main.go 装配行恒定不变）。
func NewService() *Service { return &Service{} }

// ServiceStartup Wails 生命周期钩子：A5 在此装配 store/provider 注册表/
// active map（打开 SQLite、建 map[type]构造器）。返回 error 则应用启动即败露。
func (s *Service) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	return nil
}

// ServiceShutdown Wails 生命周期钩子：A5 在此关 store、cancel 在途请求。
func (s *Service) ServiceShutdown() {}

// Ping 链路验证，返回 pong 标识。
func (s *Service) Ping() string { return "pong:tako" }

var errNotImplemented = fmt.Errorf("not implemented (SPEC-A pending)")

// Translate 发起多服务翻译（fan-out），事件经 translate:event 推送。
func (s *Service) Translate(ctx context.Context, req TextReq) (TranslateHandle, error) {
	return TranslateHandle{}, errNotImplemented
}

// CancelTranslate 取消一次进行中的翻译。
func (s *Service) CancelTranslate(requestID string) error { return errNotImplemented }

// TestService 用给定配置做连通性验证（未保存的服务也可测）。
func (s *Service) TestService(ctx context.Context, service ServiceSetting) (TestResult, error) {
	return TestResult{}, errNotImplemented
}

// ListServices 返回全部服务配置（按 order 升序）。
func (s *Service) ListServices() ([]ServiceSetting, error) {
	return nil, errNotImplemented
}

// SaveService 保存（新增或更新）服务配置。
func (s *Service) SaveService(service ServiceSetting) (ServiceSetting, error) {
	return ServiceSetting{}, errNotImplemented
}

// DeleteService 删除服务配置。
func (s *Service) DeleteService(id string) error { return errNotImplemented }

// ListHistory 分页查询历史。
func (s *Service) ListHistory(query HistoryQuery) (HistoryPage, error) {
	return HistoryPage{}, errNotImplemented
}

// SaveHistory 保存一条历史（翻译完成时调用）。
func (s *Service) SaveHistory(item HistoryItem) (HistoryItem, error) {
	return HistoryItem{}, errNotImplemented
}

// DeleteHistory 删除一条历史。
func (s *Service) DeleteHistory(id string) error { return errNotImplemented }

// ToggleFavorite 切换收藏状态。
func (s *Service) ToggleFavorite(id string, favorite bool) error { return errNotImplemented }

// ClearHistory 清空历史（收藏一并清除）。
func (s *Service) ClearHistory() error { return errNotImplemented }

// GetSettings 读取应用设置。
func (s *Service) GetSettings() (AppSettings, error) { return AppSettings{}, errNotImplemented }

// SaveSettings 保存应用设置。
func (s *Service) SaveSettings(settings AppSettings) error { return errNotImplemented }

// CheckAccessibility macOS 辅助功能权限（模拟按键取词依赖）；Windows 恒 true。
func (s *Service) CheckAccessibility() bool { return true }

// OpenAccessibilitySettings 打开系统辅助功能设置面板。
func (s *Service) OpenAccessibilitySettings() error { return errNotImplemented }

// LicenseInfo 许可证状态（一期恒 free）。
func (s *Service) LicenseInfo() LicenseInfo { return LicenseInfo{Licensed: true, Plan: "free"} }
