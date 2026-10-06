/**
 * 翻译 Store (Setup Store).
 *
 * B1 骨架：持有输入文本与语言对。流式多开聚合（requestId / 子流状态 Map /
 * onTranslateEvent 订阅）在 B3 按 provider-streaming 技能补全。
 */
import { ref } from 'vue';
import { defineStore } from 'pinia';

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

  return {
    input,
    sourceLang,
    targetLang,
  };
});
