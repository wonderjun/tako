<template>
  <!-- 透明窗口内的卡片本体 -->
  <div class="min-h-screen bg-transparent p-3">
    <div class="flex min-h-0 flex-col gap-3 rounded-xl border bg-card p-4 text-card-foreground shadow-lg">
      <!-- 顶部：服务切换 + 操作 -->
      <div class="flex items-center gap-2">
        <Segmented
          v-if="segmentItems.length > 1"
          :model-value="activeServiceId"
          :items="segmentItems"
          @update:model-value="(v: string) => (activeServiceId = v)"
        />
        <span v-else-if="segmentItems.length === 1" class="truncate text-sm font-medium">
          {{ segmentItems[0]?.label }}
        </span>
        <div class="flex-1" />
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label="复制译文"
          title="复制译文"
          :disabled="!currentText"
          @click="copyResult"
        >
          <Copy :size="14" />
        </Button>
        <Button
          variant="ghost"
          size="icon-xs"
          :aria-label="pinned ? '取消钉住' : '钉住'"
          :title="pinned ? '取消钉住' : '钉住'"
          @click="pinned = !pinned"
        >
          <Pin :size="14" :class="pinned ? 'fill-current' : ''" />
        </Button>
        <Button variant="ghost" size="icon-xs" aria-label="关闭" title="关闭" @click="onClose">
          <X :size="14" />
        </Button>
      </div>

      <!-- 译文区（复用 B3 的流式渲染） -->
      <stream-result v-if="activeServiceId" :service-id="activeServiceId" class="min-h-28" />
      <p v-else class="py-6 text-center text-sm text-muted-foreground">暂无翻译结果</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { toast } from 'vue-sonner';
import { Copy, Pin, X } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Segmented } from '@/components/ui/segmented';
import { desktop } from '@/shared/desktop/adapter';
import StreamResult from '@/features/translate/components/stream-result.vue';
import { useTranslateStore } from '@/features/translate/store';

/**
 * popup 浮窗：紧凑结果面板（透明窗口内的卡片）。
 * 监听 selection:captured（划词取词完成）→ 填入文本并自动多开翻译；
 * 钉住/关闭仅 UI 态，真实窗口控制后续由主线接 wails API。
 */
const store = useTranslateStore();

/** 当前展示的服务子流. */
const activeServiceId = ref('');

/** 钉住态（仅 UI 视觉，窗口置顶控制权后续接 wails API）. */
const pinned = ref(false);

/** selection:captured 订阅的取消函数. */
let unsubscribe: (() => void) | null = null;

/** 服务切换项（标签 = 服务名）. */
const segmentItems = computed(() => store.activeServices.map((s) => ({ value: s.id, label: s.name })));

/** 当前服务的译文文本. */
const currentText = computed(() => store.streams.get(activeServiceId.value)?.text ?? '');

onMounted(async () => {
  await store.loadServices();
  activeServiceId.value = store.activeServices[0]?.id ?? '';

  const bridge = await desktop();
  unsubscribe = bridge.onSelectionCaptured(async (p) => {
    store.input = p.text;
    await store.translate();
    activeServiceId.value = store.activeServices[0]?.id ?? '';
  });
});

onUnmounted(() => {
  unsubscribe?.();
});

/** 复制当前服务译文. */
async function copyResult(): Promise<void> {
  if (!currentText.value) return;
  try {
    await navigator.clipboard.writeText(currentText.value);
    toast.success('译文已复制');
  } catch {
    toast.error('复制失败');
  }
}

/** 关闭：仅 UI 态清理（窗口隐藏后续由主线接 wails API）. */
function onClose(): void {
  store.reset();
  activeServiceId.value = '';
}
</script>
