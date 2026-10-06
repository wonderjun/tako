<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes['class'];
    // CUSTOMIZED: 新增尺寸档 (上游无此 prop): sm→h-7 / default→h-8 / lg→h-9 (§2.7 刻度)
    size?: 'sm' | 'default' | 'lg';
  }>(),
  { size: 'default' },
);

// CUSTOMIZED: 尺寸高度走普通 h-* 类而非 data-size 选择器——普通类优先级 (0,1,0)
// 才不会反杀 has-[>textarea]:h-auto (0,1,1) 与 block-start/end 的 h-auto 覆盖
const sizeClasses = { sm: 'h-7', default: 'h-8', lg: 'h-9' } as const;
</script>

<template>
  <div
    data-slot="input-group"
    role="group"
    :class="
      cn(
        // CUSTOMIZED: v2 输入配方 — 去 shadow-xs, transition 收敛 transition-colors, h-9→h-8;
        // hover 加 border-hairline-strong; 聚焦由 border-hairline-strong+ring-2 收敛为
        // border-input-focus 单边 (上游 border-ring + ring-3); invalid 去 ring 只留 border
        'group/input-group border-input dark:bg-input/30 relative flex w-full items-center rounded-md border transition-colors outline-none',
        'hover:border-hairline-strong',
        'min-w-0 has-[>textarea]:h-auto',
        sizeClasses[props.size],

        // Variants based on alignment.
        'has-[>[data-align=inline-start]]:[&>input]:pl-2',
        'has-[>[data-align=inline-end]]:[&>input]:pr-2',
        'has-[>[data-align=block-start]]:h-auto has-[>[data-align=block-start]]:flex-col has-[>[data-align=block-start]]:[&>input]:pb-3',
        'has-[>[data-align=block-end]]:h-auto has-[>[data-align=block-end]]:flex-col has-[>[data-align=block-end]]:[&>input]:pt-3',

        // CUSTOMIZED: 聚焦收敛为 input-focus 描边 (上游 border-ring + ring-3)
        'has-[[data-slot=input-group-control]:focus-visible]:border-input-focus',

        // Error state.
        'has-[[data-slot][aria-invalid=true]]:border-destructive',

        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
