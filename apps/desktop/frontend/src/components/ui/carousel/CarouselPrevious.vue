<script setup lang="ts">
import type { WithClassAsProps } from './interface';
import type { ButtonVariants } from '@/components/ui/button';
import { ArrowLeft } from '@lucide/vue';
import { cn } from '@/shared/lib/utils';
import { Button } from '@/components/ui/button';
import { useCarousel } from './useCarousel';

const props = withDefaults(
  defineProps<
    {
      variant?: ButtonVariants['variant'];
      size?: ButtonVariants['size'];
    } & WithClassAsProps
  >(),
  {
    variant: 'outline',
    // CUSTOMIZED: size 默认 icon→icon-lg — Button v2 刻度下移 (上游 icon=36px 对应
    // 新 icon-lg); 实际渲染由下方 class size-8 钉死 32px, 两档产出相同, 改名仅为口径一致
    size: 'icon-lg',
  },
);

const { orientation, canScrollPrev, scrollPrev } = useCarousel();
</script>

<template>
  <Button
    data-slot="carousel-previous"
    :disabled="!canScrollPrev"
    :class="
      cn(
        'absolute size-8 rounded-full',
        orientation === 'horizontal'
          ? 'top-1/2 -left-12 -translate-y-1/2'
          : '-top-12 left-1/2 -translate-x-1/2 rotate-90',
        props.class,
      )
    "
    :variant="variant"
    :size="size"
    @click="scrollPrev"
  >
    <slot>
      <ArrowLeft />
      <span class="sr-only">Previous Slide</span>
    </slot>
  </Button>
</template>
