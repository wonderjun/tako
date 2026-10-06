<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { inject } from 'vue';
import { cn } from '@/shared/lib/utils';
import { DIALOG_CONTEXT_KEY } from './dialog-context';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();

// CUSTOMIZED: gap-2→gap-1.5; variant=tray/expressive 时头部落在外壳画布上:
// min-h-11 保证头部行高 ≥ × 按钮占位 (12+28+4), 否则 × 底缘压到内卡描边;
// justify-center 让标题与 ×(top-3 居中=26px) 同行对齐; pr-10 给 × 让出车道.
const ctx = inject(DIALOG_CONTEXT_KEY, undefined);
const isTrayLike = ctx?.variant === 'tray' || ctx?.variant === 'expressive';
</script>

<template>
  <div
    data-slot="dialog-header"
    :class="
      cn(
        'flex flex-col gap-1.5 text-center sm:text-left',
        isTrayLike && 'min-h-11 justify-center px-4 py-1 pr-10',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
