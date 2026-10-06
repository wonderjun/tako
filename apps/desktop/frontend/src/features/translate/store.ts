/**
 * 翻译 Store (Setup Store).
 *
 * 数据流：translate() 取启用服务建立子流状态 → bridge.translate 拿 requestId →
 * onTranslateEvent 按 (requestId, serviceId) 聚合（applyStreamEvent 纯函数归约）；
 * finish 时逐服务落 saveHistory（serviceName 冗余存）。事件订阅随 store
 * 生命周期建立一次，迟到/旧请求事件按 requestId 过滤丢弃。
 * 服务快照经 listServices 自取（features 互禁导入，不依赖 services store）。
 */
import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { ServiceSetting, StreamEvent, StreamErrorCode } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';

/** 子流状态机：idle/pending → streaming/retrying → done | error；cancelled 归 idle. */
export type StreamStatus = 'idle' | 'pending' | 'streaming' | 'retrying' | 'done' | 'error';

/** 单个服务的翻译子流状态. */
export interface StreamState {
  status: StreamStatus;
  /** delta 累积文本（finish 时被覆盖）. */
  text: string;
  /** 终态错误（cancelled 不进此字段，归 idle）. */
  error?: { code?: StreamErrorCode; message?: string };
  /** stream-retry 次数. */
  attempt?: number;
  /** finish 落历史后的条目 id（收藏回写用）. */
  historyId?: string;
  /** 结果是否已收藏. */
  favorite?: boolean;
}

/**
 * 单子流事件归约（纯函数，spec 直接驱动）.
 * 铁律：delta 累积；finish 以 e.text 覆盖累积值（防丢字）；
 * stream-retry 不是错误；cancelled 不算失败。
 */
export function applyStreamEvent(state: StreamState, e: StreamEvent): void {
  switch (e.type) {
    case 'message-start':
      break;
    case 'delta': {
      // 丢弃前导空白（个别 LLM 会先吐换行，whitespace-pre-wrap 下渲染成
      // 顶部空行）：累积仍为全空白时重置并 trimStart 新分片，中间换行不受影响
      state.text = state.text.trim() === '' ? (e.text ?? '').trimStart() : state.text + (e.text ?? '');
      state.status = 'streaming';
      break;
    }
    case 'stream-retry':
      state.status = 'retrying';
      state.attempt = e.attempt;
      break;
    case 'finish':
      state.text = (e.text ?? state.text).trim();
      state.status = 'done';
      state.attempt = undefined;
      break;
    case 'error':
      if (e.errorCode === 'cancelled') {
        state.status = 'idle';
        return;
      }
      state.status = 'error';
      state.error = { code: e.errorCode, message: e.message };
      break;
  }
}

export const useTranslateStore = defineStore('translate', () => {
  // ============================================================================
  // State
  // ============================================================================

  /** 待翻译文本. */
  const input = ref('');

  /** 源语言（'auto' 或语言代码）. */
  const sourceLang = ref('auto');

  /** 目标语言. */
  const targetLang = ref('zh-Hans');

  /** 当前请求 id；null 表示无进行中请求（事件按此过滤）. */
  const requestId = ref<string | null>(null);

  /** serviceId → 子流状态. */
  const streams = ref(new Map<string, StreamState>());

  /** 服务快照（含禁用项，展示时过滤）. */
  const services = ref<ServiceSetting[]>([]);

  /** 选中的对照渠道；null 表示全部启用服务参与. */
  const selectedIds = ref<string[] | null>(null);

  /** 最近一次请求参数快照（retry 复用；非响应式私有状态）. */
  let lastRequest: { text: string; sourceLang: string; targetLang: string } | null = null;

  // ============================================================================
  // Getters
  // ============================================================================

  /** 全部启用中的服务（渠道选择器的选项，order 升序）. */
  const enabledServices = computed(() => services.value.filter((s) => s.enabled).sort((a, b) => a.order - b.order));

  /** 参与对照的服务（启用 ∩ 渠道选择，order 升序）. */
  const activeServices = computed(() => {
    if (!selectedIds.value) return enabledServices.value;
    const picked = new Set(selectedIds.value);
    return enabledServices.value.filter((s) => picked.has(s.id));
  });

  /** 是否有子流仍在进行中. */
  const isTranslating = computed(() =>
    [...streams.value.values()].some(
      (s) => s.status === 'pending' || s.status === 'streaming' || s.status === 'retrying',
    ),
  );

  /** 是否有可展示的子流（非 idle 或已有文本；cancelled/清空后自动隐藏）. */
  const hasVisibleStreams = computed(() =>
    [...streams.value.values()].some((s) => s.status !== 'idle' || s.text.length > 0),
  );

  // ============================================================================
  // Actions
  // ============================================================================

  /** 拉取服务快照（页面挂载时调用）. */
  async function loadServices(): Promise<void> {
    const bridge = await desktop();
    services.value = await bridge.listServices();
  }

  /** 发起多开翻译：清旧态 → 建子流 → bridge.translate → 订阅聚合. */
  async function translate(): Promise<void> {
    const text = input.value.trim();
    if (!text) return;
    if (activeServices.value.length === 0) await loadServices();
    const enabled = activeServices.value;
    if (enabled.length === 0) return;

    // 清空旧状态，为每个参与服务建立 pending 子流
    streams.value = new Map(enabled.map((s) => [s.id, { status: 'pending', text: '' } as StreamState]));
    lastRequest = { text, sourceLang: sourceLang.value, targetLang: targetLang.value };

    const bridge = await desktop();
    const handle = await bridge.translate({
      text,
      sourceLang: sourceLang.value,
      targetLang: targetLang.value,
      serviceIds: enabled.map((s) => s.id),
    });
    requestId.value = handle.requestId;
    await ensureSubscribed();
  }

  /** 取消当前请求：非终态子流归 idle；晚到的旧事件按 requestId 过滤丢弃. */
  async function cancel(): Promise<void> {
    if (!requestId.value) return;
    const bridge = await desktop();
    await bridge.cancelTranslate(requestId.value);
    requestId.value = null;
    for (const state of streams.value.values()) {
      if (state.status === 'pending' || state.status === 'streaming' || state.status === 'retrying') {
        state.status = 'idle';
      }
    }
  }

  /** 单服务重译：仅传该 serviceId 发起新请求，其余子流不受影响. */
  async function retry(serviceId: string): Promise<void> {
    if (!lastRequest) return;
    streams.value.set(serviceId, { status: 'pending', text: '' });
    const bridge = await desktop();
    const handle = await bridge.translate({ ...lastRequest, serviceIds: [serviceId] });
    requestId.value = handle.requestId;
    await ensureSubscribed();
  }

  /**
   * 切换渠道是否参与对照（同时移除其旧子流，保持展示与下一次翻译一致）.
   *
   * @param id      服务 id.
   * @param checked 是否参与.
   */
  function toggleSelected(id: string, checked: boolean): void {
    const current = new Set(selectedIds.value ?? enabledServices.value.map((s) => s.id));
    if (checked) current.add(id);
    else current.delete(id);
    selectedIds.value = [...current];
    for (const key of [...streams.value.keys()]) {
      if (!current.has(key)) streams.value.delete(key);
    }
  }

  /** 清空输入与全部子流状态（页面「清空」按钮；保留渠道选择）. */
  function reset(): void {
    input.value = '';
    requestId.value = null;
    streams.value = new Map();
    lastRequest = null;
  }

  /**
   * 收藏/取消收藏该服务的结果（finish 落库后可回写）.
   *
   * @param serviceId 子流对应的服务 id.
   */
  async function toggleFavorite(serviceId: string): Promise<void> {
    const state = streams.value.get(serviceId);
    if (!state || !state.historyId) return;
    const bridge = await desktop();
    const next = !state.favorite;
    await bridge.toggleFavorite(state.historyId, next);
    state.favorite = next;
  }

  // ---- 事件订阅（私有：每 store 实例一次，随窗口生命周期） ----

  let unsubscribe: (() => void) | null = null;

  async function ensureSubscribed(): Promise<void> {
    if (unsubscribe) return;
    const bridge = await desktop();
    unsubscribe = bridge.onTranslateEvent((e) => {
      if (e.requestId !== requestId.value) return; // 迟到/旧请求事件丢弃
      const state = streams.value.get(e.serviceId);
      if (!state) return;
      applyStreamEvent(state, e);
      if (e.type === 'finish') void saveHistoryFor(e.serviceId, state);
    });
  }

  /** finish 后逐服务落历史（serviceName 冗余存），条目 id 回写子流供收藏. */
  async function saveHistoryFor(serviceId: string, state: StreamState): Promise<void> {
    if (!lastRequest) return;
    const service = services.value.find((s) => s.id === serviceId);
    try {
      const bridge = await desktop();
      const saved = await bridge.saveHistory({
        id: '',
        createdAt: Date.now(),
        sourceText: lastRequest.text,
        sourceLang: lastRequest.sourceLang,
        targetLang: lastRequest.targetLang,
        serviceId,
        serviceName: service?.name ?? serviceId,
        resultText: state.text,
        favorite: false,
      });
      state.historyId = saved.id;
      state.favorite = saved.favorite;
    } catch {
      // 历史落库失败不影响结果展示
    }
  }

  return {
    input,
    sourceLang,
    targetLang,
    requestId,
    streams,
    services,
    selectedIds,
    enabledServices,
    activeServices,
    isTranslating,
    hasVisibleStreams,
    loadServices,
    translate,
    cancel,
    retry,
    reset,
    toggleSelected,
    toggleFavorite,
  };
});
