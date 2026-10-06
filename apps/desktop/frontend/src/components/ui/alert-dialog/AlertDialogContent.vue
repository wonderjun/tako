<script setup lang="ts">
import type { AlertDialogContentEmits, AlertDialogContentProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { AlertDialogContent, AlertDialogOverlay, AlertDialogPortal, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

defineOptions({
  inheritAttrs: false,
});

// CUSTOMIZED: 新增 size prop (sm 420 / md 480, 确认弹窗规格见 §10 Confirm), 上游固定 sm:max-w-lg
const props = withDefaults(
  defineProps<
    AlertDialogContentProps & {
      class?: HTMLAttributes['class'];
      size?: 'sm' | 'md';
    }
  >(),
  { size: 'sm' },
);
const emits = defineEmits<AlertDialogContentEmits>();

const delegatedProps = reactiveOmit(props, 'class', 'size');

const sizeClasses = {
  sm: 'sm:max-w-105',
  md: 'sm:max-w-120',
} as const;

const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <AlertDialogPortal>
    <!-- CUSTOMIZED: bg-black/80 死黑 → bg-overlay + 背景模糊 (Qoder 式磨砂蒙层) -->
    <AlertDialogOverlay
      data-slot="alert-dialog-overlay"
      class="data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-overlay backdrop-blur-sm"
    />
    <AlertDialogContent
      data-slot="alert-dialog-content"
      v-bind="{ ...$attrs, ...forwarded }"
      :class="
        cn(
          // CUSTOMIZED: 表面对齐 v2 模态配方 (rounded-lg→xl, border→hairline, shadow-lg→elevation-3,
          // 去固定 sm:max-w-lg 改由 size prop 控制, 加 max-h-90vh 限高)
          'bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] max-h-[90vh] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-xl border border-hairline p-6 shadow-elevation-3 duration-200',
          sizeClasses[props.size],
          props.class,
        )
      "
    >
      <slot />
    </AlertDialogContent>
  </AlertDialogPortal>
</template>
