import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it } from 'vitest';
import { desktop } from '@/shared/desktop/adapter';
import type { HistoryItem } from '@contracts/index';
import { useHistoryStore } from './store';

/** 经 bridge 直接落库种子数据（绕过 store，专测 store 行为）. */
async function seed(item: Partial<HistoryItem>): Promise<HistoryItem> {
  const bridge = await desktop();
  return bridge.saveHistory({
    id: '',
    createdAt: Date.now(),
    sourceText: 'seed text',
    sourceLang: 'en',
    targetLang: 'zh-Hans',
    serviceId: 'svc-a',
    serviceName: '服务 A',
    resultText: '种子译文',
    favorite: false,
    ...item,
  });
}

describe('useHistoryStore（mock 桥驱动）', () => {
  beforeEach(async () => {
    setActivePinia(createPinia());
    // mock 桥单例在本文件各测试间共享，先清空历史保证用例独立
    const bridge = await desktop();
    await bridge.clearHistory();
  });

  it('load 按关键词过滤拉取第一页', async () => {
    await seed({ sourceText: 'alpha unique' });
    await seed({ sourceText: 'beta text' });

    const store = useHistoryStore();
    store.query.keyword = 'beta';
    await store.load();

    expect(store.total).toBe(1);
    expect(store.items[0]?.sourceText).toBe('beta text');
  });

  it('load 按 serviceId 与 favoriteOnly 过滤', async () => {
    await seed({ sourceText: 'from a', serviceId: 'svc-a', favorite: true });
    await seed({ sourceText: 'from b', serviceId: 'svc-b', favorite: false });

    const store = useHistoryStore();
    store.query.serviceId = 'svc-a';
    store.query.favoriteOnly = true;
    await store.load();

    expect(store.total).toBe(1);
    expect(store.items[0]?.sourceText).toBe('from a');
  });

  it('loadMore 触底追加下一页直至取完', async () => {
    // 落库 55 条（createdAt 递增保证排序稳定）
    for (let i = 0; i < 55; i++) {
      await seed({ sourceText: `bulk item ${i}`, createdAt: Date.now() + i });
    }

    const store = useHistoryStore();
    await store.load();
    expect(store.items.length).toBe(50);
    expect(store.hasMore).toBe(true);

    await store.loadMore();
    expect(store.items.length).toBe(55);
    expect(store.total).toBe(55);
    expect(store.hasMore).toBe(false);
  });

  it('query 变化后重新 load 会重置分页', async () => {
    for (let i = 0; i < 55; i++) {
      await seed({ sourceText: `reset item ${i}`, createdAt: Date.now() + i });
    }

    const store = useHistoryStore();
    await store.load();
    await store.loadMore();
    expect(store.items.length).toBe(55);

    store.query.keyword = 'reset item 54';
    await store.load();
    expect(store.items.length).toBe(1);
  });

  it('toggleFavorite 乐观翻转并经 bridge 持久化', async () => {
    const saved = await seed({ sourceText: 'fav target' });
    const store = useHistoryStore();
    await store.load();

    const item = store.items.find((i) => i.id === saved.id)!;
    await store.toggleFavorite(item);
    expect(item.favorite).toBe(true);

    const bridge = await desktop();
    const page = await bridge.listHistory({ favoriteOnly: true });
    expect(page.items.some((i) => i.id === saved.id)).toBe(true);
  });

  it('remove 删除单条并同步 total', async () => {
    const saved = await seed({ sourceText: 'to remove' });
    const store = useHistoryStore();
    await store.load();
    const before = store.total;

    await store.remove(saved.id);
    expect(store.items.some((i) => i.id === saved.id)).toBe(false);
    expect(store.total).toBe(before - 1);
  });

  it('clearAll 清空全部', async () => {
    await seed({ sourceText: 'one' });
    await seed({ sourceText: 'two' });

    const store = useHistoryStore();
    await store.load();
    expect(store.total).toBe(2);

    await store.clearAll();
    expect(store.total).toBe(0);
    expect(store.items.length).toBe(0);
  });
});
