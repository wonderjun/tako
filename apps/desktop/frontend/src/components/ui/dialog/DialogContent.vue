<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { X } from '@lucide/vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogClose, DialogContent, DialogPortal, useForwardPropsEmits } from 'reka-ui';
import { computed, provide, ref, useAttrs } from 'vue';
import { cn } from '@/shared/lib/utils';
import { DIALOG_CONTEXT_KEY } from './dialog-context';
import DialogOverlay from './DialogOverlay.vue';

defineOptions({
  inheritAttrs: false,
});

// CUSTOMIZED: v2 弹窗体系 (DESIGN.md §10) — 新增 size/variant props 并 provide 给子件;
// 表面 rounded-lg border shadow-lg sm:max-w-lg p-6 → rounded-xl border-hairline
// shadow-elevation-3 + max-h-90vh 限高, padding/宽度改由 variant/size 控制;
// 关闭按钮换成 muted 静息 + state-hover 的 28px 图标钮 (上游 ring-offset 配方).
const props = withDefaults(
  defineProps<
    DialogContentProps & {
      class?: HTMLAttributes['class'];
      showCloseButton?: boolean;
      size?: 'sm' | 'md' | 'lg' | 'xl' | 'fit' | 'full' | 'bare';
      variant?: 'compact' | 'tray' | 'expressive' | 'viewer';
    }
  >(),
  {
    showCloseButton: true,
    size: 'md',
    variant: 'compact',
  },
);
const emits = defineEmits<DialogContentEmits>();

const delegatedProps = reactiveOmit(props, 'class', 'size', 'variant');

const forwarded = useForwardPropsEmits(delegatedProps, emits);

const hasDescription = ref(false);
provide(DIALOG_CONTEXT_KEY, { variant: props.variant, size: props.size, hasDescription });

const attrs = useAttrs();
// aria-describedby 自动处理: 有 DialogDescription 时不写该属性, 交 reka 自动关联
// (显式绑 undefined 会在 mergeProps 层盖掉 reka 注入的 descriptionId, 必须连 key 一起省略);
// 无描述时显式 null 抑制告警 — 收拢业务侧手写 :aria-describedby="null" 的机制;
// 调用方显式传值 (含 null) 时以 $attrs 为准.
const describedByAttrs = computed(() => {
  if ('aria-describedby' in attrs || hasDescription.value) return {};
  return { 'aria-describedby': null };
});

const variantClasses = {
  compact: 'p-6',
  viewer: 'p-6',
  // Tray/Expressive 外壳 = 画布色 + hairline-shell 软描边 + 4px 内缩, 内卡由 DialogBody 承担
  tray: 'bg-sidebar border-hairline-shell flex flex-col gap-0 p-1',
  expressive: 'bg-sidebar border-hairline-shell flex flex-col gap-0 p-1',
} as const;

const sizeClasses = {
  sm: 'sm:max-w-105',
  md: 'sm:max-w-120',
  lg: 'sm:max-w-160',
  xl: 'sm:max-w-200',
  fit: 'sm:w-fit sm:max-w-[90vw]',
  full: 'sm:max-w-[92vw] max-h-[88vh]',
  bare: '',
} as const;

const contentClass = computed(() =>
  cn(
    // 定位 + 动画 + 限高 (全部变体共用; grid 在 tray 分支被 flex 覆盖)
    'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] max-h-[90vh] translate-x-[-50%] translate-y-[-50%] gap-4 duration-200',
    // 表面: bare 是灯箱裸壳, 整面盖掉 (不发 border/shadow 源类, 避免 shadow-elevation-3
    // 与 shadow-none 同为自定义组不被 twMerge 去重); 其余 = 纸面 + hairline + elevation-3
    props.size === 'bare'
      ? 'bg-transparent rounded-none border-none p-0 shadow-none'
      : cn('bg-background rounded-xl border border-hairline shadow-elevation-3', variantClasses[props.variant]),
    sizeClasses[props.size],
    props.class,
  ),
);
</script>

<template>
  <DialogPortal>
    <DialogOverlay />
    <DialogContent
      data-slot="dialog-content"
      v-bind="{ ...$attrs, ...forwarded, ...describedByAttrs }"
      :class="contentClass"
    >
      <slot />

      <DialogClose
        v-if="showCloseButton"
        data-slot="dialog-close"
        :class="
          cn(
            'text-muted-foreground hover:text-foreground hover:bg-state-hover absolute size-7 grid place-items-center rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 disabled:pointer-events-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*=\'size-\'])]:size-4',
            // tray/expressive: 头部 min-h-11 行高内居中 (12+14=26px) 不落进内卡;
            // 其余变体 p-6 纸面上 top-4
            props.variant === 'tray' || props.variant === 'expressive' ? 'top-3 right-3' : 'top-4 right-4',
          )
        "
      >
        <X />
        <span class="sr-only">Close</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
