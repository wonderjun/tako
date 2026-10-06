<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <th
    data-slot="table-head"
    :class="
      cn(
        // CUSTOMIZED: §9 表头 h-10 px-2 → h-8 px-3 text-xs font-medium muted
        'text-muted-foreground h-8 px-3 text-left align-middle text-xs font-medium whitespace-nowrap [&:has([role=checkbox])]:pr-0 *:[[role=checkbox]]:translate-y-0.5',
        props.class,
      )
    "
  >
    <slot />
  </th>
</template>
