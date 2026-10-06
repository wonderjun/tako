<script setup lang="ts">
import type { DialogDescriptionProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogDescription, useForwardProps } from 'reka-ui';
import { inject, onBeforeUnmount, onMounted } from 'vue';
import { cn } from '@/shared/lib/utils';
import { DIALOG_CONTEXT_KEY } from './dialog-context';

const props = defineProps<DialogDescriptionProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');

const forwardedProps = useForwardProps(delegatedProps);

// CUSTOMIZED: 描述 13px muted (§10, 上游 text-sm); 挂载时向 DialogContent 上报
// hasDescription 驱动 aria-describedby 自动关联; ctx 可能不存在 (如 command 面板
// 内复用时没有 DialogContent 包裹), 做空值守卫.
const ctx = inject(DIALOG_CONTEXT_KEY, undefined);
onMounted(() => {
  if (ctx) ctx.hasDescription.value = true;
});
onBeforeUnmount(() => {
  if (ctx) ctx.hasDescription.value = false;
});
</script>

<template>
  <DialogDescription
    data-slot="dialog-description"
    v-bind="forwardedProps"
    :class="cn('text-body text-muted-foreground', props.class)"
  >
    <slot />
  </DialogDescription>
</template>
