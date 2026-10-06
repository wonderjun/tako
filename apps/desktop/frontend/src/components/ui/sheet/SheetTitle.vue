<script setup lang="ts">
import type { DialogTitleProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogTitle } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <!-- CUSTOMIZED: 与 DialogTitle 同口径 (15px medium), 上游 'text-foreground font-semibold' -->
  <DialogTitle
    data-slot="sheet-title"
    :class="cn('text-body-lg leading-none font-medium', props.class)"
    v-bind="delegatedProps"
  >
    <slot />
  </DialogTitle>
</template>
