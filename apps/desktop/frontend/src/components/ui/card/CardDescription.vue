<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: 卡片描述 12px muted (§2.3), 上游 'text-muted-foreground text-sm' → 'text-xs text-muted-foreground' -->
  <p data-slot="card-description" :class="cn('text-xs text-muted-foreground', props.class)">
    <slot />
  </p>
</template>
