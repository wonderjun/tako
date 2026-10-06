import { describe, expect, it } from 'vitest';
import { bridgeMode, desktop } from './adapter';
import type { StreamEvent } from '@contracts/index';

describe('desktop adapter（mock 模式）', () => {
  it('纯浏览器环境回落 mock 模式', async () => {
    const bridge = await desktop();
    expect(bridgeMode()).toBe('mock');
    expect(await bridge.ping()).toBe('pong:mock');
  });

  it('mock 翻译推送 start→delta*→finish 事件序列', async () => {
    const bridge = await desktop();
    const events: StreamEvent[] = [];
    const stop = bridge.onTranslateEvent((e) => events.push(e));
    const { requestId } = await bridge.translate({
      text: 'hello tako',
      sourceLang: 'auto',
      targetLang: 'zh-Hans',
      serviceIds: ['mock-llm'],
    });
    // 等待 mock 流结束（逐 chunk 30ms × 若干片）
    await new Promise((r) => setTimeout(r, 500));
    stop();

    // 多开时同 requestId 有多条子流，按 serviceId 聚合是前端契约（contracts README）
    const mine = events.filter((e) => e.requestId === requestId && e.serviceId === 'mock-llm');
    expect(mine.at(0)?.type).toBe('message-start');
    expect(mine.at(-1)?.type).toBe('finish');
    expect(mine.at(-1)?.text).toContain('hello tako');
    const deltas = mine.filter((e) => e.type === 'delta');
    expect(deltas.length).toBeGreaterThan(0);
    // delta 拼接应等于 finish 完整文本
    expect(deltas.map((d) => d.text ?? '').join('')).toBe(mine.at(-1)?.text);
  });

  it('mock 服务与历史 CRUD 可用', async () => {
    const bridge = await desktop();
    const services = await bridge.listServices();
    expect(services.length).toBeGreaterThanOrEqual(2);

    const saved = await bridge.saveHistory({
      id: '',
      createdAt: Date.now(),
      sourceText: 'hi',
      sourceLang: 'en',
      targetLang: 'zh-Hans',
      serviceId: services[0]!.id,
      serviceName: services[0]!.name,
      resultText: '你好',
      favorite: false,
    });
    const page = await bridge.listHistory({ keyword: 'hi' });
    expect(page.total).toBe(1);
    await bridge.deleteHistory(saved.id);
    expect((await bridge.listHistory({})).total).toBe(0);
  });
});
