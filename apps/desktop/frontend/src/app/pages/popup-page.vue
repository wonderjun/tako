<template>
  <!-- 透明窗口内的卡片本体 -->
  <div class="min-h-screen bg-transparent p-3">
    <div class="flex min-h-0 flex-col gap-3 rounded-xl border bg-card p-4 text-card-foreground shadow-lg">
      <!-- 顶部：服务切换 + 操作（窄窗下 Segmented 可收缩滚动，操作按钮恒可见） -->
      <div class="flex items-center gap-2">
        <Segmented
          v-if="segmentItems.length > 1"
          :model-value="activeServiceId"
          :items="segmentItems"
          class="min-w-0 flex-1 overflow-x-auto [scrollbar-width:none]"
          @update:model-value="(v: string) => (activeServiceId = v)"
        />
        <span v-else-if="segmentItems.length === 1" class="truncate text-sm font-medium">
          {{ segmentItems[0]?.label }}
        </span>
        <div class="flex shrink-0 items-center gap-1">
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
            @click="togglePin"
          >
            <Pin :size="14" :class="pinned ? 'fill-current' : ''" />
          </Button>
          <Button variant="ghost" size="icon-xs" aria-label="关闭" title="关闭" @click="onClose">
            <X :size="14" />
          </Button>
        </div>
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
 * 钉住经 bridge.setWindowPinned 落 Go 侧 pin 态（失焦不隐藏），
 * 关闭经 bridge.hideWindow 真隐藏窗口。
 */
const store = useTranslateStore();

/** 当前展示的服务子流. */
const activeServiceId = ref('');

/** 钉住态（UI 即时反馈，失败回滚）. */
const pinned = ref(false);

/** selection:captured 订阅的取消函数. */
let unsubscribe: (() => void) | null = null;

/** 服务切换项（标签 = 服务名，窄窗下超长截断）. */
const segmentItems = computed(() =>
  store.activeServices.map((s) => ({ value: s.id, label: s.name.length > 12 ? `${s.name.slice(0, 12)}…` : s.name })),
);

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

/** 钉住/松开浮窗（置位 Go 侧 pin，失焦隐藏监听据此放行）. */
async function togglePin(): Promise<void> {
  const next = !pinned.value;
  pinned.value = next;
  try {
    const bridge = await desktop();
    await bridge.setWindowPinned('popup', next);
  } catch {
    pinned.value = !next;
    toast.error('窗口钉住失败');
  }
}

/** 关闭：真实隐藏窗口并清 UI 态. */
async function onClose(): Promise<void> {
  store.reset();
  activeServiceId.value = '';
  try {
    const bridge = await desktop();
    await bridge.hideWindow('popup');
  } catch {
    toast.error('关闭失败');
  }
}
</script>
