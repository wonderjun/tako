package bridge

import (
	"fmt"

	"tako/internal/provider"
	"tako/internal/provider/llm"
	"tako/internal/store"
)

// providerFactory 由服务运行时配置构造 provider 实例。
type providerFactory func(provider.ProviderConfig) (provider.Provider, error)

// defaultFactories 服务类型 → provider 构造器（contracts.ProviderType 全集）。
// 显式 map 而非 init() 注册：装配关系一眼可查，测试可整体替换。
func defaultFactories() map[string]providerFactory {
	return map[string]providerFactory{
		"llm": func(cfg provider.ProviderConfig) (provider.Provider, error) {
			return llm.New(cfg)
		},
		"deepl": func(cfg provider.ProviderConfig) (provider.Provider, error) {
			return provider.NewDeepL(cfg)
		},
		"google": func(cfg provider.ProviderConfig) (provider.Provider, error) {
			return provider.NewGoogle(cfg)
		},
		"microsoft": func(cfg provider.ProviderConfig) (provider.Provider, error) {
			return provider.NewMicrosoft(cfg)
		},
	}
}

// buildProvider 由 store 服务行构造 provider 实例。
func (s *Service) buildProvider(svc store.Service) (provider.Provider, error) {
	factory, ok := s.factories[svc.Type]
	if !ok {
		return nil, &provider.ProviderError{
			Code:   provider.CodeUnknown,
			Detail: fmt.Sprintf("未知服务类型: %s", svc.Type),
		}
	}
	return factory(provider.ProviderConfig{
		Type:          svc.Type,
		Name:          svc.Name,
		BaseURL:       svc.Config.BaseURL,
		APIKey:        svc.Config.APIKey,
		Model:         svc.Config.Model,
		AuthScheme:    svc.Config.AuthScheme,
		CustomHeaders: svc.Config.CustomHeaders,
	})
}

// settingToConfig bridge.ServiceSetting → provider.ProviderConfig（TestService 用：
// 未保存的服务配置也可直接构造验证连通性）。
func settingToConfig(s ServiceSetting) provider.ProviderConfig {
	return provider.ProviderConfig{
		Type:          s.Type,
		Name:          s.Name,
		BaseURL:       s.BaseURL,
		APIKey:        s.APIKey,
		Model:         s.Model,
		AuthScheme:    s.AuthScheme,
		CustomHeaders: s.CustomHeaders,
	}
}
