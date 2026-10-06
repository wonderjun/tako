<script setup lang="ts">
import { onMounted } from 'vue';
import { useSettingsStore } from '@/features/settings/store';

// 根组件：三个窗口（main/popup/input）共用本 SPA，由 hash 路由分流。
// 挂载时装配设置（含主题副作用），各窗口独立初始化各自的 store 实例。
const settingsStore = useSettingsStore();

onMounted(() => {
  void settingsStore.load();
});
</script>

<template>
  <RouterView />
</template>
