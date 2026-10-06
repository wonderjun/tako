<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <tr
    data-slot="table-row"
    :class="
      cn(
        // CUSTOMIZED: §9 行 — hover:bg-muted/50→bg-state-hover, selected:bg-muted→bg-state-selected, border-b 补 hairline
        'hover:bg-state-hover data-[state=selected]:bg-state-selected border-b border-hairline transition-colors',
        props.class,
      )
    "
  >
    <slot />
  </tr>
</template>
