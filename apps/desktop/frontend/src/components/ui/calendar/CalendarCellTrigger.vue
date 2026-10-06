<script lang="ts" setup>
import type { CalendarCellTriggerProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { CalendarCellTrigger, useForwardProps } from 'reka-ui';
import { cn } from '@/shared/lib/utils';
import { buttonVariants } from '@/components/ui/button';

const props = withDefaults(defineProps<CalendarCellTriggerProps & { class?: HTMLAttributes['class'] }>(), {
  as: 'button',
});

const delegatedProps = reactiveOmit(props, 'class');

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <CalendarCellTrigger
    data-slot="calendar-cell-trigger"
    :class="
      cn(
        buttonVariants({ variant: 'ghost' }),
        'size-8 p-0 font-normal aria-selected:opacity-100 cursor-default',
        '[&[data-today]:not([data-selected])]:bg-accent [&[data-today]:not([data-selected])]:text-accent-foreground',
        // Selected
        'data-selected:bg-primary data-selected:text-primary-foreground data-selected:opacity-100 [&[data-selected]:hover]:bg-primary data-selected:hover:text-primary-foreground data-selected:focus:bg-primary data-selected:focus:text-primary-foreground',
        // Disabled
        'data-disabled:text-muted-foreground data-disabled:opacity-50',
        // Unavailable
        // CUSTOMIZED: 不可用日期文字由 destructive-foreground 改 muted-foreground
        // (上游 destructive-foreground 在 v2 token 下是近白字, 不可用态应弱化而非警示),
        // 保留 line-through 划线.
        'data-unavailable:text-muted-foreground data-unavailable:line-through',
        // Outside months
        'data-outside-view:text-muted-foreground',
        props.class,
      )
    "
    v-bind="forwardedProps"
  >
    <slot />
  </CalendarCellTrigger>
</template>
