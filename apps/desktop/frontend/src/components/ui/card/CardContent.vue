<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

const props = defineProps<{
  class?: HTMLAttributes['class'];
}>();
</script>

<template>
  <!-- CUSTOMIZED: 去上游 px-6 (内边距收口到 Card 根 padding prop), 留空骨架透传 class -->
  <div data-slot="card-content" :class="cn(props.class)">
    <slot />
  </div>
</template>
