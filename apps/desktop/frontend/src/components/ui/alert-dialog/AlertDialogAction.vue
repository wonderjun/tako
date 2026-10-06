<script setup lang="ts">
// CUSTOMIZED: 新增 variant prop 透传 buttonVariants (上游固定走 buttonVariants() 默认档),
// 危险操作可声明 variant='destructive' 而不必堆 class 覆盖底色.
import type { AlertDialogActionProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import type { ButtonVariants } from '@/components/ui/button';
import { reactiveOmit } from '@vueuse/core';
import { AlertDialogAction } from 'reka-ui';
import { cn } from '@/shared/lib/utils';
import { buttonVariants } from '@/components/ui/button';

const props = withDefaults(
  defineProps<AlertDialogActionProps & { class?: HTMLAttributes['class']; variant?: ButtonVariants['variant'] }>(),
  { variant: 'default' },
);

const delegatedProps = reactiveOmit(props, 'class', 'variant');
</script>

<template>
  <AlertDialogAction v-bind="delegatedProps" :class="cn(buttonVariants({ variant: props.variant }), props.class)">
    <slot />
  </AlertDialogAction>
</template>
