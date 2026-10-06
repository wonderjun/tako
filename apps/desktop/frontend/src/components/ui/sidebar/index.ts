import type { VariantProps } from 'class-variance-authority';
import type { HTMLAttributes } from 'vue';
import { cva } from 'class-variance-authority';

export interface SidebarProps {
  side?: 'left' | 'right';
  variant?: 'sidebar' | 'floating' | 'inset';
  collapsible?: 'offcanvas' | 'icon' | 'none';
  class?: HTMLAttributes['class'];
}

export { default as Sidebar } from './Sidebar.vue';
export { default as SidebarContent } from './SidebarContent.vue';
export { default as SidebarFooter } from './SidebarFooter.vue';
export { default as SidebarGroup } from './SidebarGroup.vue';
export { default as SidebarGroupAction } from './SidebarGroupAction.vue';
export { default as SidebarGroupContent } from './SidebarGroupContent.vue';
export { default as SidebarGroupLabel } from './SidebarGroupLabel.vue';
export { default as SidebarHeader } from './SidebarHeader.vue';
export { default as SidebarInput } from './SidebarInput.vue';
export { default as SidebarInset } from './SidebarInset.vue';
export { default as SidebarMenu } from './SidebarMenu.vue';
export { default as SidebarMenuAction } from './SidebarMenuAction.vue';
export { default as SidebarMenuBadge } from './SidebarMenuBadge.vue';
export { default as SidebarMenuButton } from './SidebarMenuButton.vue';
export { default as SidebarMenuItem } from './SidebarMenuItem.vue';
export { default as SidebarMenuSkeleton } from './SidebarMenuSkeleton.vue';
export { default as SidebarMenuSub } from './SidebarMenuSub.vue';
export { default as SidebarMenuSubButton } from './SidebarMenuSubButton.vue';
export { default as SidebarMenuSubItem } from './SidebarMenuSubItem.vue';
export { default as SidebarProvider } from './SidebarProvider.vue';
export { default as SidebarRail } from './SidebarRail.vue';
export { default as SidebarSeparator } from './SidebarSeparator.vue';
export { default as SidebarTrigger } from './SidebarTrigger.vue';

export { useSidebar } from './utils';

// CUSTOMIZED: v2 状态层 (§2.2/§9 Sidebar) — hover/active 底色由 --sidebar-accent
// 换 --state-hover/--state-selected 统一状态层 token (文字侧 sidebar-accent-foreground
// 同步改 foreground, active 保留 font-medium).
export const sidebarMenuButtonVariants = cva(
  'peer/menu-button flex w-full items-center gap-2 overflow-hidden rounded-md p-2 text-left text-sm outline-hidden ring-sidebar-ring transition-[width,height,padding] hover:bg-state-hover hover:text-foreground focus-visible:ring-2 active:bg-state-selected active:text-foreground disabled:pointer-events-none disabled:opacity-50 group-has-data-[sidebar=menu-action]/menu-item:pr-8 aria-disabled:pointer-events-none aria-disabled:opacity-50 data-[active=true]:bg-state-selected data-[active=true]:font-medium data-[active=true]:text-foreground data-[state=open]:hover:bg-state-hover data-[state=open]:hover:text-foreground group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-2! [&>span:last-child]:truncate [&>svg]:size-4 [&>svg]:shrink-0',
  {
    variants: {
      variant: {
        default: 'hover:bg-state-hover hover:text-foreground',
        outline:
          'bg-background shadow-[0_0_0_1px_var(--sidebar-border)] hover:bg-state-hover hover:text-foreground hover:shadow-[0_0_0_1px_var(--sidebar-accent)]',
      },
      size: {
        default: 'h-8 text-sm',
        sm: 'h-7 text-xs',
        lg: 'h-12 text-sm group-data-[collapsible=icon]:p-0!',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);

export type SidebarMenuButtonVariants = VariantProps<typeof sidebarMenuButtonVariants>;
