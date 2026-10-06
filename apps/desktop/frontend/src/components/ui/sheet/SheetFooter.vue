<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{ class?: HTMLAttributes['class'] }>();
</script>

<template>
  <!-- CUSTOMIZED: 按钮排列与 DialogFooter 同口径 (flex-col-reverse + sm 右对齐), 保留 mt-auto p-4 -->
  <div
    data-slot="sheet-footer"
    :class="cn('mt-auto flex flex-col-reverse gap-2 p-4 sm:flex-row sm:justify-end', props.class)"
  >
    <slot />
  </div>
</template>
