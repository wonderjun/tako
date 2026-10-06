/**
 * 桌面桥接适配器 — 前端业务调用 Go 能力的唯一入口。
 *
 * 双模式：
 * - wails：包装 frontend/bindings（wails3 generate 产物）的 service 调用，
 *   事件经 @wailsio/runtime 的 Events 订阅
 * - mock：纯浏览器实现（见 ./mock），vite dev / vitest 下全流程可走
 *
 * 动态 import：wails 相关模块仅在真实 webview 环境加载，
 * vitest/node 下不触及。业务侧统一 `const bridge = await desktop()`
 * （幂等，初始化一次后在 store 内持有）。
 *
 * 红线：业务代码禁止直接 import bindings 或 @wailsio/runtime，
 * 一律经 desktop() 获取 DesktopBridge。
 */
import type { Bridge, StreamEvent } from '@contracts/index';
import { BridgeEvents } from '@contracts/index';
import { createEmitter, type Emitter, type Unsubscribe } from './emitter';
import { createMockBridge, type MockBridgeEvents } from './mock';

export type BridgeMode = 'wails' | 'mock';

/** Bridge 方法调用 + Go→前端事件订阅的统一面。 */
export interface DesktopBridge extends Bridge {
  onTranslateEvent(cb: (e: StreamEvent) => void): Unsubscribe;
  onSelectionCaptured(cb: (p: { text: string }) => void): Unsubscribe;
  onHotkeyTriggered(cb: (p: { kind: string }) => void): Unsubscribe;
}

/**
 * wails 环境标识：@wailsio/runtime 在 webview 注入的内部全局。
 * 纯浏览器（vite dev 直开 / vitest）无此标识 → mock 模式。
 */
function isWailsEnv(): boolean {
  return typeof window !== 'undefined' && '_wails' in window;
}

const bridgeEvents = createEmitter<MockBridgeEvents>();

// bindings 的调用面：wails3 generate 按方法名生成导出函数
type BindingsModule = Record<string, (...args: unknown[]) => Promise<unknown>>;

async function createWailsBridge(): Promise<DesktopBridge> {
  const [svcRaw, rt] = await Promise.all([
    // bindings 是 wails3 generate 的 js 产物（gitignore，make ci/dev 前置生成），
    // 无 .d.ts，运行时形态由 BindingsModule 约束
    // @ts-expect-error TS7016/TS2307: 生成物无类型声明
    import('../../../bindings/tako/internal/bridge/service.js'),
    import('@wailsio/runtime'),
  ]);
  const svc = svcRaw as unknown as BindingsModule;

  for (const [name, key] of [
    ['translateEvent', BridgeEvents.translateEvent],
    ['selectionCaptured', BridgeEvents.selectionCaptured],
    ['hotkeyTriggered', BridgeEvents.hotkeyTriggered],
  ] as const) {
    rt.Events.On(key, (data: unknown) => {
      bridgeEvents.emit(name, data as never);
    });
  }

  return {
    ping: () => svc.Ping() as Promise<string>,
    translate: (req) => svc.Translate(req) as never,
    cancelTranslate: (id) => svc.CancelTranslate(id) as never,
    testService: (s) => svc.TestService(s) as never,
    listServices: () => svc.ListServices() as never,
    saveService: (s) => svc.SaveService(s) as never,
    deleteService: (id) => svc.DeleteService(id) as never,
    listHistory: (q) => svc.ListHistory(q) as never,
    saveHistory: (i) => svc.SaveHistory(i) as never,
    deleteHistory: (id) => svc.DeleteHistory(id) as never,
    toggleFavorite: (id, f) => svc.ToggleFavorite(id, f) as never,
    clearHistory: () => svc.ClearHistory() as never,
    getSettings: () => svc.GetSettings() as never,
    saveSettings: (s) => svc.SaveSettings(s) as never,
    checkAccessibility: () => svc.CheckAccessibility() as never,
    openAccessibilitySettings: () => svc.OpenAccessibilitySettings() as never,
    licenseInfo: () => svc.LicenseInfo() as never,
    onTranslateEvent: (cb) => bridgeEvents.on('translateEvent', cb),
    onSelectionCaptured: (cb) => bridgeEvents.on('selectionCaptured', cb),
    onHotkeyTriggered: (cb) => bridgeEvents.on('hotkeyTriggered', cb),
  };
}

function createBrowserBridge(): DesktopBridge {
  const impl = createMockBridge(bridgeEvents);
  return {
    ...impl,
    onTranslateEvent: (cb) => bridgeEvents.on('translateEvent', cb),
    onSelectionCaptured: (cb) => bridgeEvents.on('selectionCaptured', cb),
    onHotkeyTriggered: (cb) => bridgeEvents.on('hotkeyTriggered', cb),
  };
}

let mode: BridgeMode | null = null;
let cached: Promise<DesktopBridge> | null = null;

/** 当前桥接模式（desktop() 首次调用后可用）。 */
export function bridgeMode(): BridgeMode | null {
  return mode;
}

/** 获取桌面桥接（幂等；首次调用探测环境并缓存 Promise）。 */
export function desktop(): Promise<DesktopBridge> {
  if (!cached) {
    mode = isWailsEnv() ? 'wails' : 'mock';
    cached = mode === 'wails' ? createWailsBridge() : Promise.resolve(createBrowserBridge());
  }
  return cached;
}

export type { Emitter };
