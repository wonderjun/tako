<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { useVModel } from '@vueuse/core';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
  defaultValue?: string | number;
  modelValue?: string | number;
}>();

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void;
}>();

const modelValue = useVModel(props, 'modelValue', emits, {
  passive: true,
  defaultValue: props.defaultValue,
});
</script>

<template>
  <!-- CUSTOMIZED: v2 输入配方 — 去 shadow-xs/ring 盒影, text-base md:text-sm→text-sm;
       hover 加 border-hairline-strong; 聚焦收敛为 border-input-focus 单边;
       invalid 去 ring-destructive 只留 border-destructive (上游 border-ring + ring-3) -->
  <textarea
    v-model="modelValue"
    data-slot="textarea"
    :class="
      cn(
        'border-input placeholder:text-muted-foreground hover:border-hairline-strong focus-visible:border-input-focus aria-invalid:border-destructive dark:bg-input/30 flex field-sizing-content min-h-16 w-full rounded-md border bg-transparent px-3 py-2 text-sm transition-colors outline-none disabled:cursor-not-allowed disabled:opacity-50',
        props.class,
      )
    "
  />
</template>
