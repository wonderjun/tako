<script setup lang="ts">
import type { TabsTriggerProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { TabsTrigger, useForwardProps } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<TabsTriggerProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <TabsTrigger
    data-slot="tabs-trigger"
    :class="
      cn(
        // CUSTOMIZED: §9 Tab 芯片 — text-sm→text-xs font-medium; active shadow-sm→shadow-elevation-1;
        // ring-3→ring-2 去 border-ring; inactive text-muted-foreground/active text-foreground (去 dark 双写 hack)
        'data-[state=active]:bg-background data-[state=active]:text-foreground focus-visible:ring-ring/50 text-muted-foreground inline-flex h-[calc(100%-1px)] flex-1 items-center justify-center gap-1.5 rounded-md border border-transparent px-2 py-1 text-xs font-medium whitespace-nowrap transition-[color,box-shadow] focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:shadow-elevation-1 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*=\'size-\'])]:size-4',
        props.class,
      )
    "
    v-bind="forwardedProps"
  >
    <slot />
  </TabsTrigger>
</template>
