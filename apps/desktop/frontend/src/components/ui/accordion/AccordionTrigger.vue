<script setup lang="ts">
import type { AccordionTriggerProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { ChevronDown } from '@lucide/vue';
import { reactiveOmit } from '@vueuse/core';
import { AccordionHeader, AccordionTrigger } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<AccordionTriggerProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <AccordionHeader class="flex">
    <AccordionTrigger
      data-slot="accordion-trigger"
      v-bind="delegatedProps"
      :class="
        cn(
          // CUSTOMIZED: §9 — 去 hover:underline 与 border-ring; py-4→py-3; 聚焦环 ring-3→ring-2
          'focus-visible:ring-ring/50 flex flex-1 items-start justify-between gap-4 rounded-md py-3 text-left text-sm font-medium transition-all outline-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50 [&[data-state=open]>svg]:rotate-180',
          props.class,
        )
      "
    >
      <slot />
      <slot name="icon">
        <ChevronDown
          class="text-muted-foreground pointer-events-none size-4 shrink-0 translate-y-0.5 transition-transform duration-200"
        />
      </slot>
    </AccordionTrigger>
  </AccordionHeader>
</template>
