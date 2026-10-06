<script setup lang="ts">
import type { AlertDialogDescriptionProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { AlertDialogDescription } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<AlertDialogDescriptionProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <!-- CUSTOMIZED: 描述 13px muted (§10), 上游 text-sm -->
  <AlertDialogDescription
    data-slot="alert-dialog-description"
    v-bind="delegatedProps"
    :class="cn('text-body text-muted-foreground', props.class)"
  >
    <slot />
  </AlertDialogDescription>
</template>
