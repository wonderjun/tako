import type { VariantProps } from 'class-variance-authority';
import { cva } from 'class-variance-authority';

export { default as Badge } from './Badge.vue';

// CUSTOMIZED: v2 徽标规格 (§9): h-5 胶囊, 去描边 (default/secondary/destructive)
// 与 aria-invalid ring、ring-3→ring-2; destructive 实底→bg-destructive-subtle 浅底;
// 新增 info/success/warning 三档浅底变体; outline 改 border-hairline.
export const badgeVariants = cva(
  'inline-flex items-center justify-center rounded-full px-2 py-0.5 text-xs font-medium h-5 w-fit whitespace-nowrap shrink-0 [&>svg]:size-3 gap-1 [&>svg]:pointer-events-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-2 transition-[color,box-shadow] overflow-hidden',
  {
    variants: {
      variant: {
        default: 'border-transparent bg-primary text-primary-foreground [a&]:hover:bg-primary/90',
        secondary: 'border-transparent bg-secondary text-secondary-foreground [a&]:hover:bg-secondary/90',
        destructive: 'border-transparent bg-destructive-subtle text-destructive [a&]:hover:bg-destructive-subtle',
        info: 'border-transparent bg-info-subtle text-info [a&]:hover:bg-info-subtle',
        success: 'border-transparent bg-success-subtle text-success [a&]:hover:bg-success-subtle',
        warning: 'border-transparent bg-warning-subtle text-warning [a&]:hover:bg-warning-subtle',
        outline: 'border-hairline text-foreground [a&]:hover:bg-accent [a&]:hover:text-accent-foreground',
      },
    },
    defaultVariants: {
      variant: 'default',
    },
  },
);
export type BadgeVariants = VariantProps<typeof badgeVariants>;
