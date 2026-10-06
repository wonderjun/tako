import { createRouter, createWebHashHistory } from 'vue-router';

/**
 * hash 路由：主窗口四页（翻译/服务/历史/设置）嵌套在 app-shell 下
 * （左侧导航 + RouterView）；popup / input 为独立窗口路由，不经壳。
 * 页面一律懒加载，路由只指各 feature 的 pages/index.vue。
 */
export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/shell/layout/app-shell.vue'),
      children: [
        { path: '', name: 'translate', component: () => import('@/features/translate/pages/index.vue') },
        { path: 'services', name: 'services', component: () => import('@/features/services/pages/index.vue') },
        { path: 'history', name: 'history', component: () => import('@/features/history/pages/index.vue') },
        { path: 'settings', name: 'settings', component: () => import('@/features/settings/pages/index.vue') },
      ],
    },
    { path: '/popup', name: 'popup', component: () => import('@/app/pages/popup-page.vue') },
    { path: '/input', name: 'input', component: () => import('@/app/pages/input-page.vue') },
  ],
});
