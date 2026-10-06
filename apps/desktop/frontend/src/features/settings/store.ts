/**
 * 应用设置 Store (Setup Store).
 *
 * 数据流：App.vue 挂载时 load()（三窗口各自实例）→ desktop().getSettings；
 * save() 经 saveSettings 落盘并回写本地权威值。主题副作用内聚在本 store：
 * settings.theme 与系统深色偏好任一变化即切换 html 的 .dark class。
 */
import { usePreferredDark } from '@vueuse/core';
import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
import type { AppSettings } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';

export const useSettingsStore = defineStore('settings', () => {
  // ============================================================================
  // State
  // ============================================================================

  /** 应用设置；null 表示尚未加载完成. */
  const settings = ref<AppSettings | null>(null);

  /** 是否正在保存. */
  const isSaving = ref(false);

  // ============================================================================
  // Actions
  // ============================================================================

  /** 加载设置（幂等，可重复调用刷新）. */
  async function load(): Promise<void> {
    const bridge = await desktop();
    settings.value = await bridge.getSettings();
  }

  /**
   * 保存设置并回写本地权威值.
   *
   * @param next 完整设置对象（由调用方基于当前值合并字段）.
   */
  async function save(next: AppSettings): Promise<void> {
    isSaving.value = true;
    try {
      const bridge = await desktop();
      await bridge.saveSettings(next);
      settings.value = next;
    } finally {
      isSaving.value = false;
    }
  }

  // ---- 主题副作用（私有，随 store 生命周期生效） ----

  /** 系统深色偏好（theme==='system' 时跟随）. */
  const prefersDark = usePreferredDark();

  /** 按 theme + 系统偏好解析并切换 html 的 .dark class. */
  function applyTheme(): void {
    // vitest/node 环境无 DOM，守卫跳过
    if (typeof document === 'undefined') return;
    const theme = settings.value?.theme ?? 'system';
    const isDark = theme === 'dark' || (theme === 'system' && prefersDark.value);
    document.documentElement.classList.toggle('dark', isDark);
  }

  watch([() => settings.value?.theme, prefersDark], applyTheme, { immediate: true });

  return {
    settings,
    isSaving,
    load,
    save,
  };
});
