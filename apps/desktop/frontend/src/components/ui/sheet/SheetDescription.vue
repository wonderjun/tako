<script setup lang="ts">
import type { DialogDescriptionProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DialogDescription } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<DialogDescriptionProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <!-- CUSTOMIZED: 与 DialogDescription 同口径 (13px muted), 上游 text-sm -->
  <DialogDescription
    data-slot="sheet-description"
    :class="cn('text-body text-muted-foreground', props.class)"
    v-bind="delegatedProps"
  >
    <slot />
  </DialogDescription>
</template>
