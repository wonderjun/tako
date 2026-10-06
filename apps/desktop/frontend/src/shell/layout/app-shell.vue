<template>
  <div class="flex h-screen overflow-hidden">
    <!-- 左侧导航（200px 画布） -->
    <aside class="flex w-[200px] shrink-0 flex-col bg-sidebar text-sidebar-foreground">
      <!-- macOS 隐形标题栏：高度须与 main.go InvisibleTitleBarHeight(50) 一致；
           左侧让出红绿灯位，标题放灯右侧垂直居中 -->
      <div class="flex h-[50px] shrink-0 items-center pl-22 pr-4">
        <span class="text-sm font-semibold tracking-wide">Tako</span>
      </div>
      <nav class="flex flex-col gap-1 px-2" aria-label="主导航">
        <RouterLink
          v-for="item in NAV_ITEMS"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-2 rounded-md px-3 py-1.5 text-body transition-colors hover:bg-state-hover hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
          exact-active-class="bg-state-selected text-foreground"
        >
          <component :is="item.icon" :size="16" />
          {{ item.label }}
        </RouterLink>
      </nav>
    </aside>
    <!-- 右侧内容区（滚动发生在内部容器，文档级不滚动） -->
    <div class="flex min-w-0 flex-1 flex-col bg-background text-foreground">
      <!-- 顶部 50px 隐形标题栏拖拽区兼页面标题栏（非滚动），与左侧 Tako 行同高；
           只放标题文本：此区可交互控件的拖拽手势会被窗口拖动抢占 -->
      <header class="flex h-[50px] shrink-0 items-center px-6">
        <h1 class="truncate text-lg font-semibold select-none">{{ route.meta.title }}</h1>
      </header>
      <main class="min-h-0 flex-1 overflow-y-auto">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
// 应用壳：左侧 200px 导航 + 顶部标题栏 + 右侧路由出口；主窗口四页共用，
// popup/input 独立窗口不经过本壳（路由层分流）。页面标题读 route.meta.title。
import { useRoute } from 'vue-router';
import { History, Languages, Plug, Settings } from '@lucide/vue';

const route = useRoute();

/** 侧栏导航项（图标 + 路由 + 文案）. */
const NAV_ITEMS = [
  { to: '/', label: '翻译', icon: Languages },
  { to: '/services', label: '服务', icon: Plug },
  { to: '/history', label: '历史', icon: History },
  { to: '/settings', label: '设置', icon: Settings },
] as const;
</script>
