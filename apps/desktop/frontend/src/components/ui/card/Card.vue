<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';
import { provideCardSurface, useCardNestingGuard } from './card-surface';

// CUSTOMIZED: v2 卡片规格 (DESIGN.md §8/§9). 上游 'rounded-xl border py-6 shadow-sm gap-6'
// → 'rounded-lg border-hairline bg-card 无阴影 p-4 gap-3'; 新增 props:
// padding (default p-4 gap-3 / compact p-3 gap-2 / none 配 divide-y), interactive
// (hover 描边加深 + 焦点环), selected (border-primary), as (换标签).
const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes['class'];
    padding?: 'default' | 'compact' | 'none';
    interactive?: boolean;
    selected?: boolean;
    as?: string;
  }>(),
  {
    padding: 'default',
    as: 'div',
  },
);

useCardNestingGuard('Card');
provideCardSurface();

const paddingClasses = {
  default: 'p-4 gap-3',
  compact: 'p-3 gap-2',
  none: '',
} as const;
</script>

<template>
  <component
    :is="as"
    data-slot="card"
    data-card-surface
    :class="
      cn(
        'flex flex-col rounded-lg border border-hairline bg-card text-card-foreground',
        paddingClasses[props.padding],
        props.interactive &&
          'transition-colors hover:border-hairline-strong focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50',
        props.selected && 'border-primary',
        props.class,
      )
    "
  >
    <slot />
  </component>
</template>
