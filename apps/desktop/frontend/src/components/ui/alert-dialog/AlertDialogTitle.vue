<script setup lang="ts">
import type { AlertDialogTitleProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { AlertDialogTitle } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<AlertDialogTitleProps & { class?: HTMLAttributes['class'] }>();

const delegatedProps = reactiveOmit(props, 'class');
</script>

<template>
  <!-- CUSTOMIZED: 弹窗标题 15px medium (§10), 上游 'text-lg font-semibold' -->
  <AlertDialogTitle
    data-slot="alert-dialog-title"
    v-bind="delegatedProps"
    :class="cn('text-body-lg leading-none font-medium', props.class)"
  >
    <slot />
  </AlertDialogTitle>
</template>
