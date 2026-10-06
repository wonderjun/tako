import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it } from 'vitest';
import { useServicesStore } from './store';

/**
 * services store 集成测试：经 desktop() 的 mock 桥驱动（真定时器，
 * 与 adapter.spec.ts 同款方式），断言 save/toggle/remove 的本地列表
 * 与 bridge 落盘一致。
 */
describe('useServicesStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it('load 拉取 mock 预置服务且按 order 升序', async () => {
    const store = useServicesStore();
    await store.load();
    expect(store.services.length).toBeGreaterThanOrEqual(2);
    const orders = store.services.map((s) => s.order);
    expect([...orders].sort((a, b) => a - b)).toEqual(orders);
  });

  it('save 新增服务时 id 由 bridge 生成且 order 取当前最大 +1', async () => {
    const store = useServicesStore();
    await store.load();
    const maxOrder = Math.max(...store.services.map((s) => s.order));
    const saved = await store.save({ id: '', name: '测试服务', type: 'google', enabled: true });
    expect(saved.id).not.toBe('');
    expect(saved.order).toBe(maxOrder + 1);
    expect(store.services.find((s) => s.id === saved.id)).toBeDefined();
  });

  it('save 编辑已有服务时保留原 order 与 createdAt', async () => {
    const store = useServicesStore();
    await store.load();
    const target = store.services[0]!;
    const saved = await store.save({
      id: target.id,
      name: '改名服务',
      type: target.type,
      enabled: target.enabled,
      order: target.order,
    });
    expect(saved.order).toBe(target.order);
    expect(saved.createdAt).toBe(target.createdAt);
    expect(store.services.find((s) => s.id === target.id)?.name).toBe('改名服务');
  });

  it('toggle 翻转 enabled 并经 bridge 持久化', async () => {
    const store = useServicesStore();
    await store.load();
    const target = store.services[0]!;
    const before = target.enabled;
    await store.toggle(target.id);
    expect(target.enabled).toBe(!before);

    // 重新拉取验证 bridge 侧已落盘
    await store.load();
    expect(store.services.find((s) => s.id === target.id)?.enabled).toBe(!before);
  });

  it('remove 从本地列表与 bridge 侧同时删除', async () => {
    const store = useServicesStore();
    await store.load();
    const saved = await store.save({ id: '', name: '待删除服务', type: 'deepl', enabled: true });
    await store.remove(saved.id);
    expect(store.services.find((s) => s.id === saved.id)).toBeUndefined();

    await store.load();
    expect(store.services.find((s) => s.id === saved.id)).toBeUndefined();
  });
});
