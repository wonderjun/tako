// CUSTOMIZED: 尺寸刻度整体下移一档 (上游 sm→本库 default 32px, 上游 default→本库 lg 36px;
// icon-sm→icon 32px, icon→icon-lg 36px, icon-lg 40px 移除; xs/icon-xs 不变);
// 新增 ghost-muted/ghost-destructive 变体; ghost/outline hover 由 bg-accent 改用 --state-hover;
// focus-visible 去 border-ring 且 ring-3→ring-2; destructive 走 --destructive-solid + --destructive-foreground,
// 去掉 dark:bg-destructive/60 与 focus-visible:ring-destructive; outline 用 border-hairline-strong 替代
// dark 分支 border-input/bg-input; aria-invalid 去掉 ring-destructive 只留 border.
// 升级映射约定: 上游 sm → 本库 default, 上游 default → 本库 lg (见 CUSTOMIZATIONS.md).
import type { VariantProps } from 'class-variance-authority';
import { cva } from 'class-variance-authority';

export { default as Button } from './Button.vue';

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:ring-ring/50 focus-visible:ring-2 aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground hover:bg-primary/90',
        destructive: 'bg-destructive-solid text-destructive-foreground hover:bg-destructive-solid/90',
        outline: 'border border-hairline-strong bg-background hover:bg-state-hover hover:text-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/75',
        ghost: 'hover:bg-state-hover hover:text-foreground',
        'ghost-muted': 'text-muted-foreground hover:bg-state-hover hover:text-foreground',
        'ghost-destructive': 'text-destructive hover:bg-destructive-subtle',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-8 gap-1.5 rounded-md px-3 has-[>svg]:px-2.5',
        xs: "h-6 gap-1 rounded-md px-2 text-xs has-[>svg]:px-1.5 [&_svg:not([class*='size-'])]:size-3",
        sm: "h-7 gap-1.5 rounded-md px-2.5 text-xs has-[>svg]:px-2 [&_svg:not([class*='size-'])]:size-3.5",
        lg: 'h-9 px-4 py-2 has-[>svg]:px-3',
        icon: 'size-8',
        'icon-xs': "size-6 rounded-md [&_svg:not([class*='size-'])]:size-3",
        'icon-sm': 'size-7',
        'icon-lg': 'size-9',
      },
    },
    defaultVariants: {
      variant: 'default',
      size: 'default',
    },
  },
);
export type ButtonVariants = VariantProps<typeof buttonVariants>;
