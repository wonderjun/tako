import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it } from 'vitest';
import { desktop } from '@/shared/desktop/adapter';
import { applyStreamEvent, useTranslateStore, type StreamState } from './store';

/** 等待 mock 流结束（30ms/片 × 若干片，取足量余量）. */
async function flushStream(ms = 800): Promise<void> {
  await new Promise((r) => setTimeout(r, ms));
}

describe('applyStreamEvent', () => {
  it('delta 累积文本并置 streaming', () => {
    const state: StreamState = { status: 'pending', text: '' };
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'delta', text: '你好' });
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'delta', text: '，世界' });
    expect(state.text).toBe('你好，世界');
    expect(state.status).toBe('streaming');
  });

  it('finish 以 e.text 覆盖累积文本（防丢字）', () => {
    const state: StreamState = { status: 'streaming', text: '累积' };
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'finish', text: '完整译文' });
    expect(state.text).toBe('完整译文');
    expect(state.status).toBe('done');
  });

  it('stream-retry 置重试态并记录次数，不渲染为失败', () => {
    const state: StreamState = { status: 'pending', text: '' };
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'stream-retry', attempt: 2 });
    expect(state.status).toBe('retrying');
    expect(state.attempt).toBe(2);
  });

  it('error(cancelled) 归 idle 而非失败态', () => {
    const state: StreamState = { status: 'streaming', text: '部分' };
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'error', errorCode: 'cancelled' });
    expect(state.status).toBe('idle');
    expect(state.error).toBeUndefined();
  });

  it('error(其他码) 置失败态并记录 code 与 message', () => {
    const state: StreamState = { status: 'streaming', text: '' };
    applyStreamEvent(state, { requestId: 'r', serviceId: 's', type: 'error', errorCode: 'auth', message: 'bad key' });
    expect(state.status).toBe('error');
    expect(state.error).toEqual({ code: 'auth', message: 'bad key' });
  });
});

describe('useTranslateStore（mock 桥驱动）', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('translate 为启用服务建立子流并聚合到完成', async () => {
    const store = useTranslateStore();
    await store.loadServices();
    store.input = 'hello tako';
    await store.translate();
    await flushStream();

    expect(store.activeServices.length).toBeGreaterThanOrEqual(2);
    for (const s of store.activeServices) {
      const state = store.streams.get(s.id);
      expect(state?.status).toBe('done');
      expect(state?.text).toBe('【mock→zh-Hans】hello tako');
    }
  });

  it('finish 后逐服务落历史（每个服务一条）', async () => {
    const store = useTranslateStore();
    await store.loadServices();
    store.input = 'history case';
    await store.translate();
    await flushStream();

    const bridge = await desktop();
    const page = await bridge.listHistory({ keyword: 'history case' });
    expect(page.total).toBe(store.activeServices.length);
    for (const item of page.items) {
      expect(item.serviceName).not.toBe('');
    }
  });

  it('toggleSelected 收窄对照渠道并同步移除旧子流', async () => {
    const store = useTranslateStore();
    await store.loadServices();
    expect(store.activeServices.length).toBe(store.enabledServices.length);

    const target = store.enabledServices[1]!;
    store.input = 'channel case';
    await store.translate();
    await flushStream();

    // 关掉第二个渠道：activeServices 收窄，其子流被移除
    store.toggleSelected(target.id, false);
    expect(store.selectedIds).not.toContain(target.id);
    expect(store.activeServices.some((s) => s.id === target.id)).toBe(false);
    expect(store.streams.has(target.id)).toBe(false);

    // 重新勾回：null 化之前的选择集合包含它
    store.toggleSelected(target.id, true);
    expect(store.selectedIds).toContain(target.id);
    expect(store.activeServices.length).toBe(store.enabledServices.length);
  });

  it('cancel 后非终态子流归 idle 且迟到事件被 requestId 过滤', async () => {
    const store = useTranslateStore();
    await store.loadServices();
    store.input = 'cancel case';
    await store.translate();
    await store.cancel();
    // mock 流在取消后仍会继续推送，晚到事件必须被丢弃
    await flushStream();

    expect(store.requestId).toBeNull();
    for (const state of store.streams.values()) {
      expect(state.status).toBe('idle');
    }
  });

  it('retry 仅重译目标服务，其余子流保持完成态', async () => {
    const store = useTranslateStore();
    await store.loadServices();
    store.input = 'retry case';
    await store.translate();
    await flushStream();

    const target = store.activeServices[0]!;
    const other = store.activeServices[1]!;
    await store.retry(target.id);
    await flushStream();

    expect(store.streams.get(target.id)?.status).toBe('done');
    expect(store.streams.get(target.id)?.text).toBe('【mock→zh-Hans】retry case');
    expect(store.streams.get(other.id)?.status).toBe('done');
    expect(store.streams.get(other.id)?.text).toBe('【mock→zh-Hans】retry case');
  });
});
