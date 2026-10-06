<!-- CUSTOMIZED: 上游 kbd 组件按 v2 规格新增 (§9: rounded-sm bg-muted px-1.5
     font-mono text-xs text-muted-foreground 无框) -->
<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <kbd
    data-slot="kbd"
    :class="
      cn('inline-flex items-center rounded-sm bg-muted px-1.5 font-mono text-xs text-muted-foreground', props.class)
    "
  >
    <slot />
  </kbd>
</template>
