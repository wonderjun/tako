<template>
  <main class="flex h-screen min-h-0 flex-col gap-3 bg-background p-4 text-foreground">
    <!-- 顶部：输入 + 语言对 + 钉住 -->
    <div class="flex items-center gap-2">
      <Input
        ref="inputComp"
        v-model="store.input"
        autofocus
        placeholder="输入要翻译的文本"
        size="sm"
        class="flex-1"
        @keydown="onKeydown"
      />
      <Select v-model="store.sourceLang">
        <SelectTrigger size="sm" class="w-28" aria-label="源语言">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="l in LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
        </SelectContent>
      </Select>
      <ArrowRight :size="14" class="shrink-0 text-muted-foreground" aria-hidden="true" />
      <Select v-model="store.targetLang">
        <SelectTrigger size="sm" class="w-28" aria-label="目标语言">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem v-for="l in TARGET_LANGUAGES" :key="l.code" :value="l.code">{{ l.label }}</SelectItem>
        </SelectContent>
      </Select>
      <Button size="sm" :disabled="!canTranslate || store.isTranslating" @click="onTranslate">翻译</Button>
      <Button
        variant="ghost"
        size="icon-xs"
        :aria-label="pinned ? '取消钉住' : '钉住'"
        :title="pinned ? '取消钉住' : '钉住'"
        @click="togglePin"
      >
        <Pin :size="14" :class="pinned ? 'fill-current' : ''" />
      </Button>
    </div>

    <!-- 渠道行：勾选参与对照的渠道 + 切换当前展示的子流 -->
    <div v-if="store.services.length > 0" class="flex items-center gap-2">
      <!-- 对照渠道多选（勾选后菜单保持打开，便于连续调整；与翻译页同模式） -->
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <Button variant="outline" size="sm" aria-label="选择对照渠道">
            <Layers :size="14" />
            渠道 {{ store.activeServices.length }}/{{ store.enabledServices.length }}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="start" class="w-56">
          <DropdownMenuLabel>对照渠道</DropdownMenuLabel>
          <DropdownMenuCheckboxItem
            v-for="s in store.enabledServices"
            :key="s.id"
            :model-value="store.selectedIds?.includes(s.id) ?? true"
            @select.prevent
            @update:model-value="(v: boolean) => store.toggleSelected(s.id, v)"
          >
            {{ s.name }}
          </DropdownMenuCheckboxItem>
        </DropdownMenuContent>
      </DropdownMenu>
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
    </div>

    <!-- 译文区（复用 B3 的流式渲染） -->
    <stream-result v-if="activeServiceId" :service-id="activeServiceId" class="min-h-0 flex-1" />
    <p v-else class="flex-1 content-center text-center text-sm text-muted-foreground">输入文本开始翻译</p>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import { ArrowRight, Layers, Pin } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Input } from '@/components/ui/input';
import { Segmented } from '@/components/ui/segmented';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { LANGUAGES } from '@contracts/index';
import { desktop } from '@/shared/desktop/adapter';
import StreamResult from '@/features/translate/components/stream-result.vue';
import { useTranslateStore } from '@/features/translate/store';

/** 目标语言选项（auto 仅用于源语言）. */
const TARGET_LANGUAGES = LANGUAGES.filter((l) => l.code !== 'auto');

/**
 * input 输入窗：输入即译面板（独立窗口，store 实例独立初始化）。
 * hotkey:triggered 唤起时聚焦输入框；Enter 触发翻译。
 * 钉住经 bridge.setWindowPinned 落 Go 侧 pin 态，失焦监听据此放行隐藏。
 */
const store = useTranslateStore();

/** 展示的服务子流. */
const activeServiceId = ref('');

/** 钉住态（UI 即时反馈，失败回滚）. */
const pinned = ref(false);

const inputComp = ref<{ focus: () => void } | null>(null);

/** hotkey:triggered 订阅的取消函数. */
let unsubscribe: (() => void) | null = null;

const canTranslate = computed(() => store.input.trim().length > 0);

/** 服务切换项（标签 = 服务名，窄窗下超长截断；与 popup 同规）. */
const segmentItems = computed(() =>
  store.activeServices.map((s) => ({ value: s.id, label: s.name.length > 12 ? `${s.name.slice(0, 12)}…` : s.name })),
);

// 渠道勾选变化后当前展示子流可能失效（被移除），回落到首个参与渠道
watch(
  () => store.activeServices.map((s) => s.id).join(','),
  () => {
    if (!store.activeServices.some((s) => s.id === activeServiceId.value)) {
      activeServiceId.value = store.activeServices[0]?.id ?? '';
    }
  },
);

onMounted(async () => {
  await store.loadServices();
  activeServiceId.value = store.activeServices[0]?.id ?? '';

  const bridge = await desktop();
  const s = await bridge.getSettings();
  store.sourceLang = s.sourceLang;
  store.targetLang = s.targetLang;

  unsubscribe = bridge.onHotkeyTriggered(() => {
    inputComp.value?.focus();
  });
});

onUnmounted(() => {
  unsubscribe?.();
});

/** Enter 触发翻译（IME 组合中不触发）. */
function onKeydown(e: KeyboardEvent): void {
  if (e.key !== 'Enter' || e.shiftKey || e.isComposing) return;
  e.preventDefault();
  if (canTranslate.value && !store.isTranslating) void onTranslate();
}

/** 翻译后若展示子流失效则回落首个渠道（沿用既有则不打断查看）. */
async function onTranslate(): Promise<void> {
  await store.translate();
  if (!store.activeServices.some((s) => s.id === activeServiceId.value)) {
    activeServiceId.value = store.activeServices[0]?.id ?? '';
  }
}

/** 钉住/松开输入窗（置位 Go 侧 pin，失焦隐藏监听据此放行）. */
async function togglePin(): Promise<void> {
  const next = !pinned.value;
  pinned.value = next;
  try {
    const bridge = await desktop();
    await bridge.setWindowPinned('input', next);
  } catch {
    pinned.value = !next;
    toast.error('窗口钉住失败');
  }
}
</script>
