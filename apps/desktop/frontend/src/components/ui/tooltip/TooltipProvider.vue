<script setup lang="ts">
import type { TooltipProviderProps } from 'reka-ui';
import { TooltipProvider } from 'reka-ui';

// CUSTOMIZED: 默认延迟 delayDuration 0→300ms 并补 skipDelayDuration 300ms
// (上游未设默认; 全局 tooltip 悬停 300ms 才出, 连续浏览时 300ms 内快速复现)
const props = withDefaults(defineProps<TooltipProviderProps>(), {
  delayDuration: 300,
  skipDelayDuration: 300,
});
</script>

<template>
  <TooltipProvider v-bind="props">
    <slot />
  </TooltipProvider>
</template>
