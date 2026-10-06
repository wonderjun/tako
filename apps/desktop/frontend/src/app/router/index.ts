import { createRouter, createWebHashHistory } from 'vue-router';

declare module 'vue-router' {
  interface RouteMeta {
    /** 主窗口标题栏文案（app-shell 顶部 50px 条渲染，features 不直引 shell）. */
    title?: string;
  }
}

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
        {
          path: '',
          name: 'translate',
          component: () => import('@/features/translate/pages/index.vue'),
          meta: { title: '翻译' },
        },
        {
          path: 'services',
          name: 'services',
          component: () => import('@/features/services/pages/index.vue'),
          meta: { title: '翻译服务' },
        },
        {
          path: 'history',
          name: 'history',
          component: () => import('@/features/history/pages/index.vue'),
          meta: { title: '历史' },
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/features/settings/pages/index.vue'),
          meta: { title: '设置' },
        },
      ],
    },
    { path: '/popup', name: 'popup', component: () => import('@/app/pages/popup-page.vue') },
    { path: '/input', name: 'input', component: () => import('@/app/pages/input-page.vue') },
  ],
});
