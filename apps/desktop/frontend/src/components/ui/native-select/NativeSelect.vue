<script setup lang="ts">
import type { AcceptableValue } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { ChevronDownIcon } from '@lucide/vue';
import { reactiveOmit, useVModel } from '@vueuse/core';
import { cn } from '@/shared/lib/utils';

defineOptions({
  inheritAttrs: false,
});

const props = withDefaults(
  defineProps<{
    modelValue?: AcceptableValue | AcceptableValue[];
    class?: HTMLAttributes['class'];
    // CUSTOMIZED: 新增尺寸档 (上游无此 prop): sm→h-7 / default→h-8 / lg→h-9 (§2.7 刻度)
    size?: 'sm' | 'default' | 'lg';
  }>(),
  { size: 'default' },
);

const emit = defineEmits<{
  'update:modelValue': AcceptableValue;
}>();

const modelValue = useVModel(props, 'modelValue', emit, {
  passive: true,
  defaultValue: '',
});

// CUSTOMIZED: size 是组件 prop 而非透传属性——原生 <select size> 是可见行数, 必须拦截
const delegatedProps = reactiveOmit(props, 'class', 'size');
</script>

<template>
  <div class="group/native-select relative w-fit has-[select:disabled]:opacity-50" data-slot="native-select-wrapper">
    <select
      v-bind="{ ...$attrs, ...delegatedProps }"
      v-model="modelValue"
      data-slot="native-select"
      :data-size="size"
      :class="
        cn(
          // CUSTOMIZED: v2 输入配方 — 去 selection 着色/shadow-xs/ring 盒影, h-9→h-8;
          // hover 加 border-hairline-strong; 聚焦收敛为 border-input-focus 单边
          // (上游 border-ring + ring-3); invalid 去 ring 只留 border; 新增 size 三档
          'border-input placeholder:text-muted-foreground dark:bg-input/30 dark:hover:bg-input/50 w-full min-w-0 appearance-none rounded-md border bg-transparent px-3 py-2 pr-9 text-sm transition-colors outline-none disabled:pointer-events-none disabled:cursor-not-allowed',
          'data-[size=sm]:h-7 data-[size=default]:h-8 data-[size=lg]:h-9',
          'hover:border-hairline-strong',
          'focus-visible:border-input-focus',
          'aria-invalid:border-destructive',
          props.class,
        )
      "
    >
      <slot />
    </select>
    <ChevronDownIcon
      class="text-muted-foreground pointer-events-none absolute top-1/2 right-3.5 size-4 -translate-y-1/2 opacity-50 select-none"
      aria-hidden="true"
      data-slot="native-select-icon"
    />
  </div>
</template>
