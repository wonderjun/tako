<script setup lang="ts">
import type { HTMLAttributes } from 'vue';
import { cn } from '@/shared/lib/utils';

interface SkeletonProps {
  class?: HTMLAttributes['class'];
}

const props = defineProps<SkeletonProps>();
</script>

<template>
  <!-- CUSTOMIZED: bg-primary/10 → bg-muted (§9, 素笺等主题下 primary 带色相不发橙) -->
  <div data-slot="skeleton" :class="cn('animate-pulse rounded-md bg-muted', props.class)" />
</template>
