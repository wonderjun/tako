<script lang="ts" setup>
import type { ToasterProps } from 'vue-sonner';
import { CircleCheckIcon, InfoIcon, OctagonXIcon, TriangleAlertIcon, XIcon } from '@lucide/vue';
import { computed } from 'vue';
import { Toaster as Sonner } from 'vue-sonner';
import { cn } from '@/shared/lib/utils';
import { Spinner } from '@/components/ui/spinner';

const props = defineProps<ToasterProps>();

// CUSTOMIZED: toast 阴影对到 elevation-2 (§9 浮层档), 与 props.toastOptions 合并
// 时后者优先 (调用方可覆写 classes.toast).
const mergedToastOptions = computed(() => ({
  ...props.toastOptions,
  classes: {
    toast: 'shadow-elevation-2',
    ...props.toastOptions?.classes,
  },
}));
</script>

<template>
  <Sonner
    :class="cn('toaster group', props.class)"
    :style="{
      '--normal-bg': 'var(--popover)',
      '--normal-text': 'var(--popover-foreground)',
      '--normal-border': 'var(--border)',
      '--border-radius': 'var(--radius)',
    }"
    v-bind="props"
    :toast-options="mergedToastOptions"
  >
    <!-- CUSTOMIZED: icon 槽位染状态色 (§2.1 形态1 文字/图标 text-s); loading 槽
         换 <Spinner> (上游手写 Loader2 animate-spin, §3 禁手写加载动画) -->
    <template #success-icon>
      <CircleCheckIcon class="size-4 text-success" />
    </template>
    <template #info-icon>
      <InfoIcon class="size-4 text-info" />
    </template>
    <template #warning-icon>
      <TriangleAlertIcon class="size-4 text-warning" />
    </template>
    <template #error-icon>
      <OctagonXIcon class="size-4 text-destructive" />
    </template>
    <template #loading-icon>
      <Spinner class="size-4" />
    </template>
    <template #close-icon>
      <XIcon class="size-4" />
    </template>
  </Sonner>
</template>
