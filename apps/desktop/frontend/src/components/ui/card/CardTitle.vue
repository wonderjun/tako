<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: 卡片标题 14px/500 (§2.3), 上游 'leading-none font-semibold' → 'text-sm leading-none font-medium' -->
  <h3 data-slot="card-title" :class="cn('text-sm leading-none font-medium', props.class)">
    <slot />
  </h3>
</template>
