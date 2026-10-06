/**
 * 纯浏览器 mock 桥接实现（SPEC-B 开发模式）。
 *
 * 在 `wails3 dev` 之外（vite dev / vitest / 纯浏览器）提供与 contracts 完全
 * 同构的行为：内存 CRUD、假流式翻译（逐字 delta + finish）、本地事件总线。
 * 让前端在不依赖 Go 后端进度的情况下走通全部 UI 流程。
 */
import type {
  AppSettings,
  Bridge,
  HistoryItem,
  HistoryPage,
  HistoryQuery,
  ServiceSetting,
  StreamEvent,
  TestResult,
  TextReq,
  TranslateHandle,
} from '@contracts/index';
import { createEmitter, type Emitter } from './emitter';

let idCounter = 0;
function nextId(prefix: string): string {
  idCounter += 1;
  return `${prefix}-${Date.now().toString(36)}-${idCounter}`;
}

function now(): number {
  return Date.now();
}

const PRESET_SERVICES: ServiceSetting[] = [
  {
    id: 'mock-llm',
    name: 'Mock LLM（OpenAI 兼容）',
    type: 'llm',
    enabled: true,
    baseUrl: 'https://api.example.com/v1',
    model: 'mock-model',
    authScheme: 'bearer',
    order: 0,
    createdAt: now(),
    updatedAt: now(),
  },
  {
    id: 'mock-deepl',
    name: 'Mock DeepL Free',
    type: 'deepl',
    enabled: true,
    order: 1,
    createdAt: now(),
    updatedAt: now(),
  },
];

export interface MockBridgeEvents {
  translateEvent: StreamEvent;
  selectionCaptured: { text: string };
  hotkeyTriggered: { kind: string };
}

export function createMockBridge(events: Emitter<MockBridgeEvents>): Bridge {
  const services = new Map(PRESET_SERVICES.map((s) => [s.id, { ...s }]));
  const history = new Map<string, HistoryItem>();

  function mockTranslateText(text: string, targetLang: string): string {
    // 假译文：回显 + 目标语言标记，模拟 LLM 逐字产出
    return `【mock→${targetLang}】${text}`;
  }

  return {
    async ping() {
      return 'pong:mock';
    },

    async translate(req: TextReq): Promise<TranslateHandle> {
      const requestId = nextId('req');
      const ids =
        req.serviceIds.length > 0 ? req.serviceIds : [...services.values()].filter((s) => s.enabled).map((s) => s.id);
      for (const serviceId of ids) {
        // 每个服务一条独立子流：start → delta* → finish
        void (async () => {
          events.emit('translateEvent', { requestId, serviceId, type: 'message-start' });
          const full = mockTranslateText(req.text, req.targetLang);
          // 逐 chunk 推送，每 30ms 一片，模拟流式
          for (const chunk of full.match(/.{1,4}/gs) ?? []) {
            await new Promise((r) => setTimeout(r, 30));
            events.emit('translateEvent', { requestId, serviceId, type: 'delta', text: chunk });
          }
          events.emit('translateEvent', {
            requestId,
            serviceId,
            type: 'finish',
            text: full,
            usage: { promptTokens: 12, completionTokens: full.length },
          });
        })();
      }
      return { requestId };
    },

    async cancelTranslate() {
      /* mock：事件流自然结束，无需处理 */
    },

    async testService(_service): Promise<TestResult> {
      await new Promise((r) => setTimeout(r, 200));
      return { ok: true, latencyMs: 200, sample: mockTranslateText('hello', 'zh-Hans') };
    },

    async listServices() {
      return [...services.values()].sort((a, b) => a.order - b.order);
    },

    async saveService(service) {
      // 新增（id 空串）由桥侧生成 id 并落 createdAt，编辑保留原 createdAt；与 saveHistory 同约定
      const id = service.id || nextId('svc');
      const existing = services.get(id);
      const saved: ServiceSetting = { ...service, id, createdAt: existing?.createdAt ?? now(), updatedAt: now() };
      services.set(id, saved);
      return saved;
    },

    async deleteService(id) {
      services.delete(id);
    },

    async listHistory(query: HistoryQuery): Promise<HistoryPage> {
      let items = [...history.values()].sort((a, b) => b.createdAt - a.createdAt);
      if (query.favoriteOnly) items = items.filter((i) => i.favorite);
      if (query.serviceId) items = items.filter((i) => i.serviceId === query.serviceId);
      if (query.keyword) {
        const kw = query.keyword.toLowerCase();
        items = items.filter((i) => i.sourceText.toLowerCase().includes(kw) || i.resultText.toLowerCase().includes(kw));
      }
      const total = items.length;
      const offset = query.offset ?? 0;
      const limit = query.limit ?? 50;
      return { items: items.slice(offset, offset + limit), total };
    },

    async saveHistory(item) {
      const saved = { ...item, id: item.id || nextId('h') };
      history.set(saved.id, saved);
      return saved;
    },

    async deleteHistory(id) {
      history.delete(id);
    },

    async toggleFavorite(id, favorite) {
      const item = history.get(id);
      if (item) item.favorite = favorite;
    },

    async clearHistory() {
      history.clear();
    },

    async getSettings(): Promise<AppSettings> {
      return {
        sourceLang: 'auto',
        targetLang: 'zh-Hans',
        hotkeys: { selectionTranslate: 'Alt+D', inputTranslate: 'Alt+A' },
        theme: 'system',
        launchAtLogin: false,
        maxConcurrent: 5,
      };
    },

    async saveSettings() {
      /* mock：无需持久化 */
    },

    async checkAccessibility() {
      return true;
    },

    async openAccessibilitySettings() {
      /* mock：无系统面板 */
    },

    async licenseInfo() {
      return { licensed: true, plan: 'free' as const };
    },
  };
}

// re-export 便于 spec 引用
export { createEmitter };
