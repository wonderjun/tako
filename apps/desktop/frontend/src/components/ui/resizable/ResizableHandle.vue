<script setup lang="ts">
import type { SplitterResizeHandleEmits, SplitterResizeHandleProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { SplitterResizeHandle, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<SplitterResizeHandleProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<SplitterResizeHandleEmits>();

const delegatedProps = reactiveOmit(props, 'class');
const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <!-- CUSTOMIZED: 移除上游 withHandle grip 图标, 分割条收敛为 1px 发丝线; hover/拖拽经 reka
       data-state 着主题 accent 色 (bg-primary, paper=陶土橘/matcha=抹茶绿…), 隐形热区加宽到 8px.
       依据: 对齐 Qoder 分割条交互 (无边框图标、拖拽线条变色).
       两档态都用 data-state 而非 CSS :hover — state 互斥, 杜绝拖拽期 hover 类与 drag 类
       同优先级决胜导致的着色闪烁. -->
  <SplitterResizeHandle
    data-slot="resizable-handle"
    v-bind="forwarded"
    :class="
      cn(
        'bg-hairline focus-visible:ring-ring relative flex w-px cursor-col-resize items-center justify-center transition-colors duration-150 after:absolute after:inset-y-0 after:left-1/2 after:w-2 after:-translate-x-1/2 data-[state=drag]:bg-primary data-[state=hover]:bg-primary/40 focus-visible:ring-1 focus-visible:ring-offset-1 focus-visible:outline-hidden data-[orientation=vertical]:h-px data-[orientation=vertical]:w-full data-[orientation=vertical]:cursor-row-resize data-[orientation=vertical]:after:left-0 data-[orientation=vertical]:after:h-2 data-[orientation=vertical]:after:w-full data-[orientation=vertical]:after:-translate-y-1/2 data-[orientation=vertical]:after:translate-x-0',
        props.class,
      )
    "
  />
</template>
