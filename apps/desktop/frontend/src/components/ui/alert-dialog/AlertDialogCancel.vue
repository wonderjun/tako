<script setup lang="ts">
// CUSTOMIZED: 取消按钮由 outline 改 secondary 变体 (v2 弹窗取消走次级灰底而非描边).
import type { AlertDialogCancelProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { AlertDialogCancel } from 'reka-ui';
import { cn } from '@/shared/lib/utils';
import { buttonVariants } from '@/components/ui/button';

const props = defineProps<AlertDialogCancelProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <AlertDialogCancel
    v-bind="delegatedProps"
    :class="cn(buttonVariants({ variant: 'secondary' }), 'mt-2 sm:mt-0', props.class)"
  >
    <slot />
  </AlertDialogCancel>
</template>
