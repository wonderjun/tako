// Package bridge 是暴露给前端的 Wails 服务层：方法名与参数结构对齐
// packages/contracts/src/index.ts（json tag camelCase 逐字段一致）。
//
// P0 为主线骨架（全部 stub）；SPEC-A A5 按 contracts 实装，
// 依赖注入字段在 ServiceStartup 中装配。
package bridge

// ServiceSetting 服务配置（contracts.ServiceSetting 的 Go 镜像）。
type ServiceSetting struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Type          string            `json:"type"` // llm | deepl | google | microsoft
	Enabled       bool              `json:"enabled"`
	BaseURL       string            `json:"baseUrl,omitempty"`
	APIKey        string            `json:"apiKey,omitempty"`
	Model         string            `json:"model,omitempty"`
	AuthScheme    string            `json:"authScheme,omitempty"`
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
	Order         int               `json:"order"`
	CreatedAt     int64             `json:"createdAt"`
	UpdatedAt     int64             `json:"updatedAt"`
}

// TestResult 连通性测试结果（contracts.TestResult）。
type TestResult struct {
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latencyMs"`
	Error     string `json:"error,omitempty"`
	Sample    string `json:"sample,omitempty"`
}

// TextReq 多服务翻译请求（contracts.TextReq）。
type TextReq struct {
	Text       string   `json:"text"`
	SourceLang string   `json:"sourceLang"`
	TargetLang string   `json:"targetLang"`
	ServiceIDs []string `json:"serviceIds"`
}

// TranslateHandle 翻译调用句柄（contracts.TranslateHandle）。
type TranslateHandle struct {
	RequestID string `json:"requestId"`
}

// HistoryItem / HistoryQuery / HistoryPage 历史（contracts 同名）。
type HistoryItem struct {
	ID          string `json:"id"`
	CreatedAt   int64  `json:"createdAt"`
	SourceText  string `json:"sourceText"`
	SourceLang  string `json:"sourceLang"`
	TargetLang  string `json:"targetLang"`
	ServiceID   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	ResultText  string `json:"resultText"`
	Favorite    bool   `json:"favorite"`
}

type HistoryQuery struct {
	Keyword      string `json:"keyword,omitempty"`
	ServiceID    string `json:"serviceId,omitempty"`
	FavoriteOnly bool   `json:"favoriteOnly,omitempty"`
	Limit        int    `json:"limit,omitempty"`
	Offset       int    `json:"offset,omitempty"`
}

type HistoryPage struct {
	Items []HistoryItem `json:"items"`
	Total int           `json:"total"`
}

// HotkeyConfig / AppSettings 设置（contracts 同名）。
type HotkeyConfig struct {
	SelectionTranslate string `json:"selectionTranslate"`
	InputTranslate     string `json:"inputTranslate"`
}

type AppSettings struct {
	SourceLang    string       `json:"sourceLang"`
	TargetLang    string       `json:"targetLang"`
	Hotkeys       HotkeyConfig `json:"hotkeys"`
	Theme         string       `json:"theme"` // light | dark | system
	LaunchAtLogin bool         `json:"launchAtLogin"`
	MaxConcurrent int          `json:"maxConcurrent"`
}

// LicenseInfo 许可证状态（contracts.LicenseInfo）。
type LicenseInfo struct {
	Licensed  bool   `json:"licensed"`
	Plan      string `json:"plan"` // free | pro | trial
	ExpiresAt int64  `json:"expiresAt,omitempty"`
}
