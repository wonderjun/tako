<script setup lang="ts">
import type { DropdownMenuSubContentEmits, DropdownMenuSubContentProps } from 'reka-ui';
import type { HTMLAttributes } from 'vue';
import { reactiveOmit } from '@vueuse/core';
import { DropdownMenuSubContent, useForwardPropsEmits } from 'reka-ui';
import { cn } from '@/shared/lib/utils';

const props = defineProps<DropdownMenuSubContentProps & { class?: HTMLAttributes['class'] }>();
const emits = defineEmits<DropdownMenuSubContentEmits>();

const delegatedProps = reactiveOmit(props, 'class');

const forwarded = useForwardPropsEmits(delegatedProps, emits);
</script>

<template>
  <DropdownMenuSubContent
    data-slot="dropdown-menu-sub-content"
    v-bind="forwarded"
    :class="
      cn(
        // CUSTOMIZED: v2 浮层配方 (§9): rounded-md border shadow-lg min-w-[8rem]
        // → rounded-lg border-hairline shadow-elevation-2 min-w-40 (子菜单与主菜单同阴影档)
        'bg-popover text-popover-foreground data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2 z-50 min-w-40 origin-(--reka-dropdown-menu-content-transform-origin) overflow-hidden rounded-lg border border-hairline p-1 shadow-elevation-2',
        props.class,
      )
    "
  >
    <slot />
  </DropdownMenuSubContent>
</template>
