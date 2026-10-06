<script setup lang="ts">
import type { HTMLAttributes, Ref } from 'vue';
import { defaultDocument, useEventListener, useMediaQuery, useVModel } from '@vueuse/core';
import { TooltipProvider } from 'reka-ui';
import { computed, ref } from 'vue';
import { cn } from '@/shared/lib/utils';
import {
  provideSidebarContext,
  SIDEBAR_COOKIE_MAX_AGE,
  SIDEBAR_COOKIE_NAME,
  SIDEBAR_KEYBOARD_SHORTCUT,
  SIDEBAR_WIDTH,
  SIDEBAR_WIDTH_ICON,
} from './utils';

const props = withDefaults(
  defineProps<{
    defaultOpen?: boolean;
    open?: boolean;
    class?: HTMLAttributes['class'];
    /** CUSTOMIZED: 是否禁用 Cmd/Ctrl + B 默认折叠快捷键。 */
    disableToggleShortcut?: boolean;
  }>(),
  {
    defaultOpen: !defaultDocument?.cookie.includes(`${SIDEBAR_COOKIE_NAME}=false`),
    open: undefined,
    disableToggleShortcut: false,
  },
);

const emits = defineEmits<{
  'update:open': [open: boolean];
}>();

const isMobile = useMediaQuery('(max-width: 768px)');
const openMobile = ref(false);

const open = useVModel(props, 'open', emits, {
  defaultValue: props.defaultOpen ?? false,
  passive: (props.open === undefined) as false,
}) as Ref<boolean>;

function setOpen(value: boolean) {
  open.value = value; // emits('update:open', value)

  // This sets the cookie to keep the sidebar state.
  document.cookie = `${SIDEBAR_COOKIE_NAME}=${open.value}; path=/; max-age=${SIDEBAR_COOKIE_MAX_AGE}`;
}

function setOpenMobile(value: boolean) {
  openMobile.value = value;
}

// Helper to toggle the sidebar.
function toggleSidebar() {
  return isMobile.value ? setOpenMobile(!openMobile.value) : setOpen(!open.value);
}

useEventListener('keydown', (event: KeyboardEvent) => {
  // CUSTOMIZED: 允许业务层关闭 shadcn-vue 默认的 Cmd/Ctrl + B 折叠快捷键。
  if (props.disableToggleShortcut) return;
  if (event.key === SIDEBAR_KEYBOARD_SHORTCUT && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    toggleSidebar();
  }
});

// We add a state so that we can do data-state="expanded" or "collapsed".
// This makes it easier to style the sidebar with Tailwind classes.
const state = computed(() => (open.value ? 'expanded' : 'collapsed'));

provideSidebarContext({
  state,
  open,
  setOpen,
  isMobile,
  openMobile,
  setOpenMobile,
  toggleSidebar,
});
</script>

<template>
  <!-- CUSTOMIZED: 侧栏 tooltip 延迟 0→300ms + skipDelay 300ms (与全局 TooltipProvider 默认对齐) -->
  <TooltipProvider :delay-duration="300" :skip-delay-duration="300">
    <div
      data-slot="sidebar-wrapper"
      :style="{
        '--sidebar-width': SIDEBAR_WIDTH,
        '--sidebar-width-icon': SIDEBAR_WIDTH_ICON,
      }"
      :class="cn('group/sidebar-wrapper has-data-[variant=inset]:bg-sidebar flex min-h-svh w-full', props.class)"
      v-bind="$attrs"
    >
      <slot />
    </div>
  </TooltipProvider>
</template>
