<template>
  <main class="flex h-screen min-h-0 flex-col gap-3 bg-background p-4 text-foreground">
    <!-- 顶部：输入 + 语言对 -->
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
      <Button size="sm" :disabled="!canTranslate || store.isTranslating" @click="store.translate()">翻译</Button>
    </div>

    <!-- 译文区（复用 B3 的流式渲染） -->
    <stream-result v-if="activeServiceId" :service-id="activeServiceId" class="min-h-0 flex-1" />
    <p v-else class="flex-1 content-center text-center text-sm text-muted-foreground">输入文本开始翻译</p>
  </main>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { ArrowRight } from '@lucide/vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
 */
const store = useTranslateStore();

/** 展示的服务子流（首个启用服务）. */
const activeServiceId = ref('');

const inputComp = ref<{ focus: () => void } | null>(null);

/** hotkey:triggered 订阅的取消函数. */
let unsubscribe: (() => void) | null = null;

const canTranslate = computed(() => store.input.trim().length > 0);

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
  if (canTranslate.value && !store.isTranslating) void store.translate();
}
</script>
