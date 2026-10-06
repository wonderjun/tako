<script setup lang="ts">
import type { AccordionItemProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { AccordionItem, useForwardProps } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<AccordionItemProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <!-- CUSTOMIZED: 项间分隔线补 border-hairline (§9) -->
  <AccordionItem
    v-slot="slotProps"
    data-slot="accordion-item"
    v-bind="forwardedProps"
    :class="cn('border-b border-hairline last:border-b-0', props.class)"
  >
    <slot v-bind="slotProps" />
  </AccordionItem>
</template>
