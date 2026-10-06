import type { ClassValue } from 'clsx';
import { clsx } from 'clsx';
import { extendTailwindMerge } from 'tailwind-merge';

// text-body / text-body-lg 是自定义正文字号档 (--text-body 13px / --text-body-lg 15px):
// 注册进 theme.text 使其归入 font-size 组 — 与 text-sm 等标准字号互相合并
// (后者胜), 但不再被误认成文字颜色而与 text-warning/text-foreground 冲突吞掉.
const twMerge = extendTailwindMerge({
  extend: {
    theme: {
      text: ['body', 'body-lg'],
    },
  },
});

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
