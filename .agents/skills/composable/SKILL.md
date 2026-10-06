---
name: composable
description: 在 tako 前端 抽取或编写 composable（use-*.ts，数据 hooks / 副作用 hooks / 全局单例）时使用。
---

# Composable Conventions

composable 是逻辑复用的唯一手段（不写类、不写 mixin）。本文骨架提炼自 `src/features/collab/pages/composables/use-spaces.ts`（数据 hooks）与 `src/features/image-gen/pages/composables/use-task-polling.ts`（副作用 hooks）。

## 适用场景

- 多个组件要共享一段响应式逻辑（拉数据、轮询、滚动同步、快捷键）。
- 给 Vue Query 包一层领域 hooks。
- 判断一段逻辑该不该抽成 composable、放哪。

## 两种形态（必须对号入座）

**形态一：数据 hooks** —— 返回 Vue Query 结果，调用方解构使用：

```ts
/** 空间列表（includeArchived 支持响应式引用，切换时自动重取）. */
export function useCollabSpaces(includeArchived: MaybeRefOrGetter<boolean> = false) {
  return useQuery({
    queryKey: [...COLLAB_SPACES_QUERY_KEY, computed(() => toValue(includeArchived))],
    queryFn: () => listCollabSpaces(toValue(includeArchived)),
    staleTime: 30_000,
  });
}
```

**形态二：副作用 hooks** —— 返回 `void`，内部自转（轮询、事件同步、watch 回写 store）：

```ts
/**
 * (文件头块注释写全数据流: 驱动源、回写目标、退避与兜底规则)
 */
export function useImageGenTaskPolling(): void {
  const store = useImageGenStore();
  const pendingTaskIds = computed(() => store.history.filter((t) => t.status === 'pending').map((t) => t.taskId));
  const results = useQueries({ queries: computed(() => pendingTaskIds.value.map((taskId) => ({ /* ... */ }))) });
  watch(results, (rs) => { /* 回写 store */ }, { deep: true });
}
```

规则：

1. 文件名 `use-xxx.ts`、导出 `useXxx`，一个文件一簇同域 hooks（`use-spaces.ts` 装下 collab 空间全部 hooks）。
2. **文件头块注释必写数据流**（谁驱动、写向哪、边界规则），参考 use-task-polling 的 15 行头注释。
3. 响应式入参钦点 `MaybeRefOrGetter` + `toValue`（或 `() => T` getter + `computed`），**不收裸 ref 类型**；可选响应式参数给默认值（`= false`）。
4. 数据 hooks 内的 query/mutation 写法、toast、失效全部遵循 `api-query`；副作用 hooks 回写 store 遵循 `pinia-store`。
5. 条件启停用 `enabled: computed(() => !!x())`，不在 hooks 里手写 if 装卸 watcher。
6. 组合多个简单 hooks 的"聚合 hooks"允许返回对象（如 `use-todo-data.ts` 返回 `{ items, isLoading, ... }`），但成员仍须是各 hooks 原样返回值。

## 位置

| 作用域 | 位置 | 例子 |
|---|---|---|
| 单页面私有 | `pages/composables/` | `pages/composables/use-message-scroll.ts` |
| 子域共用 | 子域 `composables/` | `todo/schedule/composables/` |
| 跨 feature | `src/shared/lib/` | `shared/lib/use-theme.ts` |

禁止新建顶层 `@/composables` 目录（tech-split 已废除）。

## 全局单例（createGlobalState）

需要跨组件共享**同一份实例状态**的全局逻辑（主题、通知）必须用 vueuse `createGlobalState`，否则多实例互不同步：

```ts
/**
 * 全局主题模式 (浅色 / 深色).
 *
 * createGlobalState 保证全局单例: 多个 useColorMode 实例在同窗口内互不同步,
 * 会出现设置页改了、侧栏状态覆盖回去的"无法点击"假象.
 */
export const useTheme = createGlobalState(() => {
  const themeMode = useColorMode<ThemeMode>({ storageKey: 'tako-theme', initialValue: 'light' });
  /* ... */
  return { themeMode, toggleDark, setTheme };
});
```

文件头必须写"为什么必须单例"（use-theme.ts 正典）；普通 hooks（每次调用新实例）不得套 createGlobalState。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| hooks 收裸 `Ref<T>` 入参 | — | 钦点 `MaybeRefOrGetter`/getter，兼容 computed 与普通值 |
| 数据 hooks 里裸写 fetch/axios | — | 请求只走 `<域>.api.ts`（api-query） |
| 该单例的不单例（直接 `export function useXxx()` 存全局态） | — | 多实例状态不同步，见 use-theme 头注释案例 |
| 组件间靠 props 层层透传而非抽 hooks | — | 三层以上透传必须抽 composable 收敛 |

## 自查清单

- [ ] 对号入座了两种形态（数据 / 副作用）？
- [ ] 文件头有数据流块注释、导出 useXxx、文件 use-xxx.ts？
- [ ] 响应式入参用 MaybeRefOrGetter + toValue？
- [ ] 位置放对（pages/composables、子域、shared/lib）？
- [ ] 全局态走了 createGlobalState 并写了单例理由？

## 范本文件

- 数据 hooks：`src/features/collab/pages/composables/use-spaces.ts`
- 副作用 hooks：`src/features/image-gen/pages/composables/use-task-polling.ts`
- 全局单例：`src/shared/lib/use-theme.ts`
- 聚合 hooks：`src/features/todo/pages/composables/use-todo-data.ts`

相关技能：`api-query`、`pinia-store`、`feature-organization`、`code-style-guide`。
