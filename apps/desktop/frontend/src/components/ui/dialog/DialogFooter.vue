<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { DialogClose } from 'reka-ui';
import { inject } from 'vue';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/components/ui/button';
import { DIALOG_CONTEXT_KEY } from './dialog-context';

// CUSTOMIZED: 内置 Close 按钮 outline→secondary (§10 纸面取消走次级灰底);
// variant=tray/expressive 时 footer 落在外壳画布上, 补 px-1 py-3 (底边距把按钮
// 抬离画布边缘, 与头部画布间隙对称);
// variant=compact 时 footer 按钮等宽 min-w-20 (§10 Compact 规格).
const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes['class'];
    showCloseButton?: boolean;
  }>(),
  {
    showCloseButton: false,
  },
);

const ctx = inject(DIALOG_CONTEXT_KEY, undefined);
const isTrayLike = ctx?.variant === 'tray' || ctx?.variant === 'expressive';
const isCompact = !ctx?.variant || ctx.variant === 'compact';
</script>

<template>
  <div
    data-slot="dialog-footer"
    :class="
      cn(
        'flex flex-col-reverse gap-2 sm:flex-row sm:justify-end',
        isTrayLike && 'px-1 py-3',
        isCompact && '[&>button]:min-w-20',
        props.class,
      )
    "
  >
    <slot />
    <DialogClose v-if="showCloseButton" as-child>
      <Button variant="secondary"> Close </Button>
    </DialogClose>
  </div>
</template>
