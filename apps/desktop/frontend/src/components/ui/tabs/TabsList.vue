<script setup lang="ts">
import type { TabsListProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { TabsList } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<TabsListProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <TabsList
    data-slot="tabs-list"
    v-bind="delegatedProps"
    :class="
      cn(
        // CUSTOMIZED: §9 Tab 轨道 h-9 p-0.75 → h-8 p-0.5 (rounded-lg bg-muted 保留)
        'bg-muted text-muted-foreground inline-flex h-8 w-fit items-center justify-center rounded-lg p-0.5',
        props.class,
      )
    "
  >
    <slot />
  </TabsList>
</template>
