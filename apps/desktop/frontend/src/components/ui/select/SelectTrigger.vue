<script setup lang="ts">
import type { SelectTriggerProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { ChevronDown } from '@lucide/vue';
import { reactiveOmit } from '@vueuse/core';
import { SelectIcon, SelectTrigger, useForwardProps } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

// CUSTOMIZED: size 扩为三档 (上游 sm/default): 新增 lg→h-9; default 由 h-9 降 h-8, sm 由 h-8 降 h-7
const props = withDefaults(
  defineProps<SelectTriggerProps & { class?: HTMLAttributes['class']; size?: 'sm' | 'default' | 'lg' }>(),
  { size: 'default' },
);

const delegatedProps = reactiveOmit(props, 'class', 'size');
const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <!-- CUSTOMIZED: v2 输入配方 — 去 shadow-xs/ring 盒影; hover 加 border-hairline-strong;
       聚焦收敛为 border-input-focus 单边; invalid 去 ring 只留 border (上游 border-ring + ring-3);
       尺寸档 data-[size=default]:h-8 / data-[size=sm]:h-7 / 新增 data-[size=lg]:h-9 -->
  <SelectTrigger
    data-slot="select-trigger"
    :data-size="size"
    v-bind="forwardedProps"
    :class="
      cn(
        'border-input data-placeholder:text-muted-foreground [&_svg:not([class*=\'text-\'])]:text-muted-foreground hover:border-hairline-strong focus-visible:border-input-focus aria-invalid:border-destructive dark:bg-input/30 dark:hover:bg-input/50 flex w-fit items-center justify-between gap-2 rounded-md border bg-transparent px-3 py-2 text-sm whitespace-nowrap transition-colors outline-none disabled:cursor-not-allowed disabled:opacity-50 data-[size=default]:h-8 data-[size=sm]:h-7 data-[size=lg]:h-9 *:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:flex *:data-[slot=select-value]:items-center *:data-[slot=select-value]:gap-2 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*=\'size-\'])]:size-4',
        props.class,
      )
    "
  >
    <slot />
    <SelectIcon as-child>
      <ChevronDown class="size-4 opacity-50" />
    </SelectIcon>
  </SelectTrigger>
</template>
