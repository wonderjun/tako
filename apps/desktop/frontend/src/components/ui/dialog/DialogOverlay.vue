<script setup lang="ts">
import type { DialogOverlayProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogOverlay } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<DialogOverlayProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <DialogOverlay
    data-slot="dialog-overlay"
    v-bind="delegatedProps"
    :class="
      cn(
        // CUSTOMIZED: bg-black/80 死黑 → bg-overlay + 背景模糊 (Qoder 式磨砂蒙层)
        'data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 fixed inset-0 z-50 bg-overlay backdrop-blur-sm',
        props.class,
      )
    "
  >
    <slot />
  </DialogOverlay>
</template>
