import type { VariantProps } from 'class-variance-authority';
import { cva } from 'class-variance-authority';

export { default as Alert } from './Alert.vue';
export { default as AlertDescription } from './AlertDescription.vue';
export { default as AlertTitle } from './AlertTitle.vue';

// CUSTOMIZED: v2 提示块规格 (§9): 只填充不加框 — 去上游 border 与 bg-card 纸底;
// default→bg-muted, destructive→bg-destructive-subtle + 图标标题 text-destructive +
// 正文 text-foreground/80; 新增 info/success/warning 三档同形态浅底变体.
export const alertVariants = cva(
  'relative w-full rounded-lg px-4 py-3 text-sm grid has-[>svg]:grid-cols-[calc(var(--spacing)*4)_1fr] grid-cols-[0_1fr] has-[>svg]:gap-x-3 gap-y-0.5 items-start [&>svg]:size-4 [&>svg]:translate-y-0.5 [&>svg]:text-current',
  {
    variants: {
      variant: {
        default: 'bg-muted text-foreground',
        destructive:
          'bg-destructive-subtle text-destructive [&>svg]:text-destructive *:data-[slot=alert-description]:text-foreground/80',
        info: 'bg-info-subtle text-info [&>svg]:text-info *:data-[slot=alert-description]:text-foreground/80',
        success:
          'bg-success-subtle text-success [&>svg]:text-success *:data-[slot=alert-description]:text-foreground/80',
        warning:
          'bg-warning-subtle text-warning [&>svg]:text-warning *:data-[slot=alert-description]:text-foreground/80',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);

export type AlertVariants = VariantProps<typeof alertVariants>;
