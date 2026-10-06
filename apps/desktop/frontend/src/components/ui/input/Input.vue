<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
// CUSTOMIZED: 加 ref / onMounted / useTemplateRef 以支持 autofocus 与对外暴露 focus/select
import { onMounted, useTemplateRef } from 'vue';
import { useVModel } from '@vueuse/core';
import { cn } from '@/shared/lib/utils';

const props = withDefaults(
  defineProps<{
    defaultValue?: string | number;
    modelValue?: string | number;
    class?: HTMLAttributes['class'];
    // CUSTOMIZED: 新增可选 autofocus prop，挂载后自动聚焦
    autofocus?: boolean;
    // CUSTOMIZED: 新增尺寸档 (上游仅默认 h-9): sm→h-7 / default→h-8 / lg→h-9
    size?: 'sm' | 'default' | 'lg';
  }>(),
  { size: 'default' },
);

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void;
}>();

const modelValue = useVModel(props, 'modelValue', emits, {
  passive: true,
  defaultValue: props.defaultValue,
});

// CUSTOMIZED: 暴露内部 input 引用，供父级模板 ref 与 onMounted 聚焦使用
const inputRef = useTemplateRef<HTMLInputElement>('inputRef');

onMounted(() => {
  if (props.autofocus) {
    inputRef.value?.focus();
  }
});

// CUSTOMIZED: 对外暴露 focus()/select()，便于父级在 nextTick 后手动聚焦/全选
defineExpose({
  focus: () => inputRef.value?.focus(),
  select: () => inputRef.value?.select(),
});
</script>

<template>
  <input
    ref="inputRef"
    v-model="modelValue"
    data-slot="input"
    :data-size="size"
    :class="
      cn(
        // CUSTOMIZED: v2 输入配方 — 去 selection 着色/shadow-xs/ring 盒影, text-base md:text-sm→text-sm;
        // 高度默认 h-9→h-8 并接 data-size 三档; hover 加 border-hairline-strong;
        // 聚焦由 border+ring 收敛为 border-input-focus 单边 (上游 border-ring + ring-3)
        'file:text-foreground placeholder:text-muted-foreground dark:bg-input/30 border-input w-full min-w-0 rounded-md border bg-transparent px-3 py-1 text-sm transition-colors outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50',
        'data-[size=sm]:h-7 data-[size=default]:h-8 data-[size=lg]:h-9',
        'hover:border-hairline-strong',
        'focus-visible:border-input-focus',
        'aria-invalid:border-destructive',
        props.class,
      )
    "
  />
</template>
