<script setup lang="ts">
import type { DialogTitleProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogTitle, useForwardProps } from 'reka-ui';
import { inject } from 'vue';
import { cn } from '@/shared/lib/utils';
import { DIALOG_CONTEXT_KEY } from './dialog-context';

const props = defineProps<DialogTitleProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');

const forwardedProps = useForwardProps(delegatedProps);

// CUSTOMIZED: 弹窗标题字号按 variant 分档 (§10): compact/viewer 15px medium,
// tray/expressive 外壳小标题 14px medium; 上游 'text-lg font-semibold' 18px/600
const ctx = inject(DIALOG_CONTEXT_KEY, undefined);
const isTrayLike = ctx?.variant === 'tray' || ctx?.variant === 'expressive';
</script>

<template>
  <DialogTitle
    data-slot="dialog-title"
    v-bind="forwardedProps"
    :class="cn(isTrayLike ? 'text-sm leading-none font-medium' : 'text-body-lg leading-none font-medium', props.class)"
  >
    <slot />
  </DialogTitle>
</template>
