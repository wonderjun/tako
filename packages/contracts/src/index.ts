/**
 * @tako/contracts — 前后端桥接契约（单一来源）
 *
 * 本包是 Wails Go 壳与 Vue 前端之间唯一的类型契约，双方都必须以此为准：
 * - Go 侧（internal/bridge）暴露的服务方法，参数/返回值的 json tag 与此处字段一一对应（camelCase）
 * - 前端（src/shared/desktop/adapter.ts）只消费此处类型，不自定义重复结构
 *
 * 红线：本包仅主线分支可修改；需要变更时在分支 PR 中提「契约变更申请」。
 */

// ==============================================================================
// 翻译服务
// ==============================================================================

/** 服务类型：llm = OpenAI 兼容；其余为传统翻译 API（非流式） */
export type ProviderType = 'llm' | 'deepl' | 'google' | 'microsoft'

/** 服务配置（存储与传输同构；apiKey 由 Go 侧写入本地 SQLite，不落前端持久化） */
export interface ServiceSetting {
  id: string
  name: string
  type: ProviderType
  enabled: boolean
  /** llm 专用 */
  baseUrl?: string
  apiKey?: string
  model?: string
  authScheme?: 'bearer' | 'x-api-key'
  customHeaders?: Record<string, string>
  /** 多开对照排序，小的在前 */
  order: number
  createdAt: number
  updatedAt: number
}

/** 连通性测试结果 */
export interface TestResult {
  ok: boolean
  latencyMs: number
  /** 失败原因（用户可读） */
  error?: string
  /** 成功时的样例译文 */
  sample?: string
}

// ==============================================================================
// 翻译请求与流式事件
// ==============================================================================

export interface TextReq {
  text: string
  /** 'auto' 或语言代码（见 LANGUAGES） */
  sourceLang: string
  targetLang: string
  /** 参与本次翻译的服务（多开对照）；空数组由 Go 侧取全部 enabled 服务 */
  serviceIds: string[]
}

/** 翻译调用返回；事件经 BridgeEvents.translateEvent 推送 */
export interface TranslateHandle {
  requestId: string
}

export type StreamEventType =
  | 'message-start'
  | 'delta'
  | 'stream-retry'
  | 'finish'
  | 'error'

/** 错误码（与 Go internal/provider.ErrorKind 对应） */
export type StreamErrorCode =
  | 'auth'
  | 'rate_limit'
  | 'timeout'
  | 'connect'
  | 'context_overflow'
  | 'network'
  | 'cancelled'
  | 'unknown'

export interface StreamEvent {
  requestId: string
  serviceId: string
  type: StreamEventType
  /** delta: 增量文本；finish: 完整译文 */
  text?: string
  /** error: 用户可读信息 */
  message?: string
  /** error 终态错误码 */
  errorCode?: StreamErrorCode
  /** stream-retry: 第几次重试（从 1 开始） */
  attempt?: number
  usage?: { promptTokens?: number; completionTokens?: number }
}

// ==============================================================================
// 历史/收藏
// ==============================================================================

export interface HistoryItem {
  id: string
  /** epoch ms */
  createdAt: number
  sourceText: string
  sourceLang: string
  targetLang: string
  serviceId: string
  serviceName: string
  resultText: string
  favorite: boolean
}

export interface HistoryQuery {
  /** 关键词模糊匹配 sourceText/resultText */
  keyword?: string
  serviceId?: string
  favoriteOnly?: boolean
  limit?: number
  offset?: number
}

export interface HistoryPage {
  items: HistoryItem[]
  total: number
}

// ==============================================================================
// 设置与许可证
// ==============================================================================

export interface HotkeyConfig {
  /** 划词翻译，默认 'Alt+D' */
  selectionTranslate: string
  /** 输入翻译，默认 'Alt+A' */
  inputTranslate: string
}

export interface AppSettings {
  sourceLang: string
  targetLang: string
  hotkeys: HotkeyConfig
  theme: 'light' | 'dark' | 'system'
  launchAtLogin: boolean
  /** 多开对照上限（1-10，默认 5） */
  maxConcurrent: number
}

export interface LicenseInfo {
  licensed: boolean
  plan: 'free' | 'pro' | 'trial'
  /** epoch ms，无期限时省略 */
  expiresAt?: number
}

// ==============================================================================
// 语言表（BCP-47 简化代码）
// ==============================================================================

export interface LanguageDef {
  code: string
  name: string
  /** 中文界面显示名 */
  label: string
}

export const LANGUAGES: readonly LanguageDef[] = [
  { code: 'auto', name: 'Auto', label: '自动检测' },
  { code: 'zh-Hans', name: 'Chinese (Simplified)', label: '简体中文' },
  { code: 'zh-Hant', name: 'Chinese (Traditional)', label: '繁體中文' },
  { code: 'en', name: 'English', label: '英语' },
  { code: 'ja', name: 'Japanese', label: '日语' },
  { code: 'ko', name: 'Korean', label: '韩语' },
  { code: 'fr', name: 'French', label: '法语' },
  { code: 'de', name: 'German', label: '德语' },
  { code: 'es', name: 'Spanish', label: '西班牙语' },
  { code: 'pt', name: 'Portuguese', label: '葡萄牙语' },
  { code: 'it', name: 'Italian', label: '意大利语' },
  { code: 'ru', name: 'Russian', label: '俄语' },
  { code: 'ar', name: 'Arabic', label: '阿拉伯语' },
  { code: 'th', name: 'Thai', label: '泰语' },
  { code: 'vi', name: 'Vietnamese', label: '越南语' },
  { code: 'id', name: 'Indonesian', label: '印尼语' },
] as const

export function languageLabel(code: string): string {
  return LANGUAGES.find((l) => l.code === code)?.label ?? code
}

// ==============================================================================
// Bridge 方法名（Go service 方法 ↔ 前端调用的唯一真源）
// ==============================================================================

/** Wails bindings 方法名 = Go (service) 方法名（导出方法，PascalCase） */
export const BridgeMethods = {
  ping: 'Ping',
  translate: 'Translate',
  cancelTranslate: 'CancelTranslate',
  testService: 'TestService',
  listServices: 'ListServices',
  saveService: 'SaveService',
  deleteService: 'DeleteService',
  listHistory: 'ListHistory',
  saveHistory: 'SaveHistory',
  deleteHistory: 'DeleteHistory',
  toggleFavorite: 'ToggleFavorite',
  clearHistory: 'ClearHistory',
  getSettings: 'GetSettings',
  saveSettings: 'SaveSettings',
  checkAccessibility: 'CheckAccessibility',
  openAccessibilitySettings: 'OpenAccessibilitySettings',
  licenseInfo: 'LicenseInfo',
  setWindowPinned: 'SetWindowPinned',
  hideWindow: 'HideWindow',
} as const

/** Go → 前端事件名（Wails application events） */
export const BridgeEvents = {
  /** payload: StreamEvent */
  translateEvent: 'translate:event',
  /** payload: { text: string } — 划词取词完成，浮窗窗口监听后展示并翻译 */
  selectionCaptured: 'selection:captured',
  /** payload: { kind: 'input' } — 输入窗唤起 */
  hotkeyTriggered: 'hotkey:triggered',
} as const

// ==============================================================================
// Bridge 接口（adapter.ts 实现的调用面；Go 侧按同名方法对齐）
// ==============================================================================

export interface Bridge {
  ping(): Promise<string>
  /** 发起翻译（多开 fan-out），事件流经 translateEvent 推送 */
  translate(req: TextReq): Promise<TranslateHandle>
  cancelTranslate(requestId: string): Promise<void>
  /** 用给定配置（未保存也可）做一次最小翻译验证连通性 */
  testService(service: ServiceSetting): Promise<TestResult>
  listServices(): Promise<ServiceSetting[]>
  saveService(service: ServiceSetting): Promise<ServiceSetting>
  deleteService(id: string): Promise<void>
  listHistory(query: HistoryQuery): Promise<HistoryPage>
  saveHistory(item: HistoryItem): Promise<HistoryItem>
  deleteHistory(id: string): Promise<void>
  toggleFavorite(id: string, favorite: boolean): Promise<void>
  clearHistory(): Promise<void>
  getSettings(): Promise<AppSettings>
  saveSettings(settings: AppSettings): Promise<void>
  /** macOS 辅助功能权限（模拟按键取词依赖）；Windows 恒 true */
  checkAccessibility(): Promise<boolean>
  openAccessibilitySettings(): Promise<void>
  licenseInfo(): Promise<LicenseInfo>
  /**
   * 钉住窗口：pinned 后失焦不自动隐藏（对 input/popup 这类
   * 失焦即隐窗才有意义；main 不失焦隐藏，调用无副作用）。
   * windowName 为 main.go 注册的窗口名（'input' | 'popup'）。
   */
  setWindowPinned(windowName: string, pinned: boolean): Promise<void>
  /** 隐藏窗口（popup 关闭按钮等）；窗口名不存在时报错 */
  hideWindow(windowName: string): Promise<void>
}
