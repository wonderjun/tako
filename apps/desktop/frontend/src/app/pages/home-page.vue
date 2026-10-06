<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { bridgeMode, desktop } from '@/shared/desktop/adapter';

/**
 * P0 首页占位：验证 bridge 链路（ping）。SPEC-B B3 将重构为
 * 翻译验证页（输入区 + 多开对照 + 流式渲染）。
 */
const pong = ref('');
const mode = ref('');

onMounted(async () => {
  const bridge = await desktop();
  mode.value = bridgeMode() ?? '';
  pong.value = await bridge.ping();
});
</script>

<template>
  <main class="flex min-h-screen items-center justify-center bg-background text-foreground">
    <div class="space-y-2 text-center">
      <h1 class="text-2xl font-semibold">Tako · 章鱼</h1>
      <p class="text-sm text-muted-foreground">
        bridge mode: <code class="font-mono">{{ mode }}</code> · ping:
        <code class="font-mono">{{ pong }}</code>
      </p>
      <p class="text-xs text-muted-foreground">P0 骨架页 · SPEC-B 将在此展开翻译界面</p>
    </div>
  </main>
</template>
