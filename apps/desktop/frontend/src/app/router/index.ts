import { createRouter, createWebHashHistory } from 'vue-router';

/**
 * hash 路由：Wails 窗口 URL 以 /#/ 前缀指定页面（main=/
 * popup=/popup input=/input）。SPEC-B 在 /services /history /settings
 * 占位页上展开实现。
 */
export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../pages/home-page.vue'),
    },
    {
      path: '/popup',
      name: 'popup',
      component: () => import('../pages/popup-page.vue'),
    },
    {
      path: '/input',
      name: 'input',
      component: () => import('../pages/input-page.vue'),
    },
    // ── SPEC-B B1 展开的占位路由 ──
    { path: '/services', name: 'services', component: () => import('../pages/home-page.vue') },
    { path: '/history', name: 'history', component: () => import('../pages/home-page.vue') },
    { path: '/settings', name: 'settings', component: () => import('../pages/home-page.vue') },
  ],
});
