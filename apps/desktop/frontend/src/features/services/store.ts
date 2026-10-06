/**
 * 翻译服务 Store (Setup Store).
 *
 * 数据流：页面挂载 load() → desktop().listServices（按 order 排序）；
 * toggle/save/remove 为命令型写操作，乐观更新本地列表后经 bridge 落盘，
 * 失败回滚并向上抛出（反馈由组件层 toast）。
 */
import { ref } from 'vue';
import { defineStore } from 'pinia';
import type { ServiceSetting } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';

/** 服务编辑草稿：时间戳由 bridge 侧维护，order 新增时自动分配. */
export type ServiceDraft = Omit<ServiceSetting, 'createdAt' | 'updatedAt' | 'order'> & { order?: number };

export const useServicesStore = defineStore('services', () => {
  // ============================================================================
  // State
  // ============================================================================

  /** 服务列表（按 order 升序）. */
  const services = ref<ServiceSetting[]>([]);

  /** 是否正在加载. */
  const isLoading = ref(false);

  // ============================================================================
  // Actions
  // ============================================================================

  /** 加载服务列表. */
  async function load(): Promise<void> {
    isLoading.value = true;
    try {
      const bridge = await desktop();
      services.value = await bridge.listServices();
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * 切换服务启用态（乐观更新，失败回滚）.
   *
   * @param id 服务 id.
   */
  async function toggle(id: string): Promise<void> {
    const target = services.value.find((s) => s.id === id);
    if (!target) return;
    const nextEnabled = !target.enabled;
    target.enabled = nextEnabled;
    try {
      const bridge = await desktop();
      const saved = await bridge.saveService({ ...target, enabled: nextEnabled });
      replace(saved);
    } catch (error) {
      target.enabled = !nextEnabled;
      throw error;
    }
  }

  /**
   * 保存服务（新增或编辑）；新增时 order 取当前最大 +1.
   *
   * @param draft 服务草稿（新增时 id 传空串）.
   * @returns bridge 落盘后的完整服务配置.
   */
  async function save(draft: ServiceDraft): Promise<ServiceSetting> {
    const bridge = await desktop();
    const isNew = !draft.id;
    const order = isNew ? Math.max(-1, ...services.value.map((s) => s.order)) + 1 : (draft.order ?? 0);
    const saved = await bridge.saveService({ ...draft, order } as ServiceSetting);
    replace(saved);
    return saved;
  }

  /**
   * 删除服务.
   *
   * @param id 服务 id.
   */
  async function remove(id: string): Promise<void> {
    const bridge = await desktop();
    await bridge.deleteService(id);
    services.value = services.value.filter((s) => s.id !== id);
  }

  // ---- 私有 helper ----

  /** 以 bridge 返回值替换列表中的同 id 项，不存在则插入（保持 order 序）. */
  function replace(saved: ServiceSetting): void {
    const index = services.value.findIndex((s) => s.id === saved.id);
    if (index >= 0) {
      services.value.splice(index, 1, saved);
    } else {
      services.value.push(saved);
      services.value.sort((a, b) => a.order - b.order);
    }
  }

  return {
    services,
    isLoading,
    load,
    toggle,
    save,
    remove,
  };
});
