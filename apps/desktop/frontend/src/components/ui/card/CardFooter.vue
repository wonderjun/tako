<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: 去上游 px-6 与 [.border-t]:pt-6 补偿 (内边距收口到 Card 根) -->
  <div data-slot="card-footer" :class="cn('flex items-center', props.class)">
    <slot />
  </div>
</template>
