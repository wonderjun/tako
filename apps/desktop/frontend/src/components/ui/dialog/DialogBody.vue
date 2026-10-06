<script setup lang="ts">
// CUSTOMIZED: 本项目新增文件 (非上游), §10 弹窗正文区: 通用 = min-h-0 滚动 +
// -ml-1 pl-1 pr-1 焦点出血惯例; variant=tray/expressive 时升级为外壳画布上的内卡
// (bg-background + border-hairline-shell, 与外壳 10/14 同心圆角, 自带 p-4 内边距,
// 内边距内聚焦 ring 不被 overflow 裁剪故不再需要出血三件套).
import type { HTMLAttributes } from 'vue';
import { inject } from 'vue';
import { cn } from '@/shared/lib/utils';
import { provideCardSurface } from '@/components/ui/card/card-surface';
import { DIALOG_CONTEXT_KEY } from './dialog-context';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();

const ctx = inject(DIALOG_CONTEXT_KEY, undefined);
const isTrayLike = ctx?.variant === 'tray' || ctx?.variant === 'expressive';

// Tray 内卡本身是卡片级容器 (L1), 其内再出现卡片由守卫拦截; compact/viewer 弹窗
// 纸面是 Frame 不是卡片, 不标记 (§8: 卡片可直接放框架面上).
if (isTrayLike) {
  provideCardSurface();
}
</script>

<template>
  <div
    data-slot="dialog-body"
    :class="
      cn(
        'min-h-0 flex-1 -ml-1 overflow-y-auto pl-1 pr-1',
        isTrayLike && 'ml-0 rounded-lg border border-hairline-shell bg-background p-4',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
