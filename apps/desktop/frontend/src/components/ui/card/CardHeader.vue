<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: v2 卡片内边距收口到 Card 根 (padding prop), 上游 px-6/[.border-b]:pb-6
       补偿与 @container/card-header 一并移除; gap-1.5→gap-1 -->
  <div
    data-slot="card-header"
    :class="
      cn(
        'grid auto-rows-min grid-rows-[auto_auto] items-start gap-1 has-data-[slot=card-action]:grid-cols-[1fr_auto]',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
