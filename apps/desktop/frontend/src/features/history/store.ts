/**
 * 翻译历史 Store (Setup Store).
 *
 * 数据流：页面挂载 load() → desktop().listHistory（query 过滤 + limit/offset
 * 分页）；query 变化由页面防抖后调 load() 重置翻页；触底 loadMore() 追加下一页。
 * 收藏/删除/清空为命令型写操作，成功后同步本地列表。
 */
import { computed, ref } from 'vue';
import { defineStore } from 'pinia';
import type { HistoryItem } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';

/** 每页条数（SPEC-B 钦点 50）. */
const PAGE_SIZE = 50;

/** 历史过滤条件（分页参数由 store 内部管理）. */
export interface HistoryFilter {
  /** 关键词（匹配原文/译文）. */
  keyword?: string;
  /** 按服务过滤. */
  serviceId?: string;
  /** 仅看收藏. */
  favoriteOnly?: boolean;
}

export const useHistoryStore = defineStore('history', () => {
  // ============================================================================
  // State
  // ============================================================================

  /** 当前页已加载的历史条目（createdAt 降序）. */
  const items = ref<HistoryItem[]>([]);

  /** 过滤条件下的总条数. */
  const total = ref(0);

  /** 过滤条件（不含分页参数）. */
  const query = ref<HistoryFilter>({});

  /** 是否正在加载. */
  const isLoading = ref(false);

  /** 分页偏移（非响应式私有状态）. */
  let offset = 0;

  // ============================================================================
  // Getters
  // ============================================================================

  /** 是否还有下一页. */
  const hasMore = computed(() => items.value.length < total.value);

  // ============================================================================
  // Actions
  // ============================================================================

  /** 按当前过滤条件加载第一页（重置分页）. */
  async function load(): Promise<void> {
    offset = 0;
    isLoading.value = true;
    try {
      const bridge = await desktop();
      const page = await bridge.listHistory({ ...query.value, offset: 0, limit: PAGE_SIZE });
      items.value = page.items;
      total.value = page.total;
    } finally {
      isLoading.value = false;
    }
  }

  /** 加载下一页并追加（触底调用；无更多或加载中时忽略）. */
  async function loadMore(): Promise<void> {
    if (isLoading.value || !hasMore.value) return;
    offset += PAGE_SIZE;
    isLoading.value = true;
    try {
      const bridge = await desktop();
      const page = await bridge.listHistory({ ...query.value, offset, limit: PAGE_SIZE });
      items.value.push(...page.items);
      total.value = page.total;
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * 切换收藏态（乐观更新，失败回滚）.
   *
   * @param item 目标条目.
   */
  async function toggleFavorite(item: HistoryItem): Promise<void> {
    const next = !item.favorite;
    item.favorite = next;
    try {
      const bridge = await desktop();
      await bridge.toggleFavorite(item.id, next);
    } catch (error) {
      item.favorite = !next;
      throw error;
    }
  }

  /**
   * 删除单条历史.
   *
   * @param id 条目 id.
   */
  async function remove(id: string): Promise<void> {
    const bridge = await desktop();
    await bridge.deleteHistory(id);
    items.value = items.value.filter((i) => i.id !== id);
    total.value = Math.max(0, total.value - 1);
  }

  /** 清空全部历史（需页面层二次确认后调用）. */
  async function clearAll(): Promise<void> {
    const bridge = await desktop();
    await bridge.clearHistory();
    items.value = [];
    total.value = 0;
    offset = 0;
  }

  return {
    items,
    total,
    query,
    isLoading,
    hasMore,
    load,
    loadMore,
    toggleFavorite,
    remove,
    clearAll,
  };
});
