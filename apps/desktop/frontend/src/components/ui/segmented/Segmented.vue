<!-- CUSTOMIZED: 本项目自研组件(非 shadcn-vue 官方组件), 勿用 CLI 覆盖; 贴内容胶囊槽: 选中白芯片(图标+文字), 有图标的未选项只露图标; locked 项追加锁徽标(放行与否由消费方在 update:model-value 拦截) -->
<script setup lang="ts" generic="T extends string">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { Lock } from '@lucide/vue';
import { cn } from '@/shared/lib/utils';

interface SegmentedItem<T> {
  value: T;
  label: string;
  icon?: any;
  /** 锁定位: 追加锁徽标, 控件不拦截切换, 由消费方在 update:model-value 自行拒绝. */
  locked?: boolean;
}

const props = withDefaults(
  defineProps<{
    modelValue: T;
    items: SegmentedItem<T>[];
    class?: string;
  }>(),
  {
    class: '',
  },
);

const emit = defineEmits<{
  'update:modelValue': [value: T];
}>();

const containerRef = ref<HTMLElement | null>(null);
const indicatorStyle = ref<{ width: string; transform: string; opacity: number }>({
  width: '0px',
  transform: 'translateX(0)',
  opacity: 0,
});
let resizeObserver: ResizeObserver | null = null;

const activeIndex = computed(() => props.items.findIndex((item) => item.value === props.modelValue));

async function updateIndicator() {
  await nextTick();
  const container = containerRef.value;
  if (!container) {
    indicatorStyle.value = {
      width: '0px',
      transform: 'translateX(0)',
      opacity: 0,
    };
    return;
  }

  const buttons = container.querySelectorAll<HTMLElement>('[data-segmented-trigger]');
  const activeBtn = buttons[activeIndex.value];
  if (!activeBtn) {
    indicatorStyle.value = {
      width: '0px',
      transform: 'translateX(0)',
      opacity: 0,
    };
    return;
  }

  indicatorStyle.value = {
    width: `${activeBtn.offsetWidth}px`,
    transform: `translateX(${activeBtn.offsetLeft}px)`,
    opacity: 1,
  };
}

watch(() => props.modelValue, updateIndicator, { immediate: true });
watch(() => props.items, updateIndicator, { deep: true });

onMounted(() => {
  updateIndicator();
  window.addEventListener('resize', updateIndicator);

  if (typeof ResizeObserver !== 'undefined' && containerRef.value) {
    resizeObserver = new ResizeObserver(() => {
      updateIndicator();
    });
    resizeObserver.observe(containerRef.value);
  }
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateIndicator);
  resizeObserver?.disconnect();
  resizeObserver = null;
});

function select(value: T) {
  if (value === props.modelValue) return;
  emit('update:modelValue', value);
}
</script>

<template>
  <div
    ref="containerRef"
    :class="
      cn(
        // CUSTOMIZED: 轨道对齐 §9 胶囊规格 bg-muted (去 bg-muted/70 透明度与 border-hairline — 填充/描边二选一 §8)
        'relative inline-flex w-auto shrink-0 self-start items-center rounded-full bg-muted p-0.5',
        props.class,
      )
    "
    role="radiogroup"
  >
    <div
      class="absolute top-0.5 left-0 z-0 h-[calc(100%-0.25rem)] rounded-full bg-card shadow-elevation-1 transition-all duration-200 ease-out"
      :style="{
        width: indicatorStyle.width,
        transform: indicatorStyle.transform,
        opacity: indicatorStyle.opacity,
      }"
    />

    <button
      v-for="item in items"
      :key="item.value"
      data-segmented-trigger
      type="button"
      role="radio"
      :aria-checked="item.value === props.modelValue"
      :aria-label="item.label"
      :title="item.label"
      :class="
        cn(
          'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 relative z-10 flex items-center justify-center gap-1.5 whitespace-nowrap rounded-full py-1 text-sm font-medium transition-[color,padding] duration-200',
          item.icon && item.value !== props.modelValue ? 'px-2' : 'px-2.5',
          item.value === props.modelValue ? 'text-foreground' : 'text-muted-foreground hover:text-foreground',
        )
      "
      @click="select(item.value)"
    >
      <component :is="item.icon" v-if="item.icon" class="size-4 shrink-0" />
      <span v-if="!item.icon || item.value === props.modelValue">{{ item.label }}</span>
      <Lock v-if="item.locked" class="size-3 shrink-0 text-muted-foreground" />
    </button>
  </div>
</template>
