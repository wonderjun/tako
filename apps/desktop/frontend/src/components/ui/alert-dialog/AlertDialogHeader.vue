<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: gap-2→gap-1.5 与 DialogHeader 对齐 -->
  <div data-slot="alert-dialog-header" :class="cn('flex flex-col gap-1.5 text-center sm:text-left', props.class)">
    <slot />
  </div>
</template>
