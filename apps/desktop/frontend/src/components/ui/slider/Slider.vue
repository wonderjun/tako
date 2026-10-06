<script setup lang="ts">
import type { SliderRootEmits, SliderRootProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { SliderRange, SliderRoot, SliderThumb, SliderTrack, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<SliderRootProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<SliderRootEmits>();

const delegatedProps = reactiveOmit(props, 'class');
const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <SliderRoot
    v-bind="forwarded"
    :class="
      cn(
        'relative flex w-full touch-none items-center select-none data-[orientation=vertical]:h-full data-[orientation=vertical]:w-auto data-[orientation=vertical]:flex-col',
        props.class,
      )
    "
  >
    <!-- CUSTOMIZED: 轨道/滑块/焦点环由硬编码 slate 改为主题 token (bg-muted/bg-primary/ring-ring 等), 适配品牌色与 dark 模式; 焦点环不透明度统一为 ring-ring/50 -->
    <SliderTrack
      class="relative h-1.5 w-full grow overflow-hidden rounded-full bg-muted data-[orientation=vertical]:w-1.5"
    >
      <SliderRange class="absolute h-full bg-primary data-[orientation=vertical]:w-full" />
    </SliderTrack>
    <SliderThumb
      v-for="(_, index) in forwarded.modelValue ?? [0]"
      :key="index"
      class="block size-4 rounded-full border-2 border-primary bg-background shadow ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
    />
  </SliderRoot>
</template>
