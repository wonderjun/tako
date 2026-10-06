<template>
  <div class="flex min-h-0 flex-1 flex-col gap-2">
    <!-- 状态区 -->
    <div class="flex min-h-5 items-center gap-2">
      <template v-if="state">
        <Spinner v-if="state.status === 'pending'" class="size-3.5 text-muted-foreground" />
        <span v-else-if="state.status === 'streaming'" class="text-xs text-muted-foreground">翻译中…</span>
        <Badge v-else-if="state.status === 'retrying'" variant="warning">重试中 第{{ state.attempt ?? 0 }}次</Badge>
      </template>
    </div>

    <!-- 错误提示 -->
    <Alert v-if="state?.status === 'error'" variant="destructive">
      <CircleAlert />
      <AlertTitle>翻译失败</AlertTitle>
      <AlertDescription>{{ errorCopy(state.error?.code, state.error?.message) }}</AlertDescription>
    </Alert>

    <!-- 译文区（流式渐进 + 末尾光标；结果可滚动） -->
    <div
      class="min-h-24 flex-1 overflow-y-auto whitespace-pre-wrap rounded-md border border-hairline bg-muted/40 p-3 text-body leading-relaxed"
    >
      <template v-if="state?.text">{{ state.text }}</template>
      <span v-else-if="state" class="text-muted-foreground">等待输出…</span>
      <span v-if="state?.status === 'streaming'" class="animate-pulse" aria-hidden="true">▍</span>
    </div>

    <!-- 操作条 -->
    <div class="flex items-center gap-1">
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="复制译文"
        title="复制译文"
        :disabled="!state?.text"
        @click="copyResult"
      >
        <Copy :size="16" />
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="重译"
        title="重译"
        :disabled="isBusy || !state"
        @click="store.retry(serviceId)"
      >
        <RotateCcw :size="16" />
      </Button>
      <Button
        variant="ghost"
        size="icon-sm"
        aria-label="收藏"
        title="收藏"
        :disabled="state?.status !== 'done' || !state?.historyId"
        @click="toggleFavorite"
      >
        <Star :size="16" :class="state?.favorite ? 'fill-current' : ''" />
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { toast } from 'vue-sonner';
import { CircleAlert, Copy, RotateCcw, Star } from '@lucide/vue';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Spinner } from '@/components/ui/spinner';
import { useTranslateStore } from '@/features/translate/store';
import { errorCopy } from '../pages/utils/error-copy';

const props = defineProps<{
  /** 子流对应的服务 id（translate store streams 的键）. */
  serviceId: string;
}>();

const store = useTranslateStore();

/** 当前子流状态. */
const state = computed(() => store.streams.get(props.serviceId));

/** 是否有子流进行中（进行中禁止重译）. */
const isBusy = computed(
  () => store.isTranslating || ['pending', 'streaming', 'retrying'].includes(state.value?.status ?? ''),
);

/** 复制当前译文. */
async function copyResult(): Promise<void> {
  if (!state.value?.text) return;
  try {
    await navigator.clipboard.writeText(state.value.text);
    toast.success('译文已复制');
  } catch {
    toast.error('复制失败');
  }
}

/** 收藏/取消收藏. */
async function toggleFavorite(): Promise<void> {
  try {
    await store.toggleFavorite(props.serviceId);
    if (state.value?.favorite) toast.success('已收藏到历史');
  } catch {
    toast.error('操作失败，请重试');
  }
}
</script>
