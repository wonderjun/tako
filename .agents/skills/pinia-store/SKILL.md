---
name: pinia-store
description: 在 tako 前端 新建或修改 Pinia store、划分客户端状态与服务端状态边界时使用。
---

# Pinia Store Conventions

**全部 setup store**（函数式 defineStore），分节注释固定三段。本文骨架提炼自 `src/features/image-gen/stores/image-gen.ts`（结构正典）。

## 适用场景

- 新建或修改客户端状态 store。
- 拿不准一段状态该进 Pinia 还是 Vue Query。
- 评审 store 的结构与命名。

## 分工铁律（先于一切）

- **服务端状态（请求来、会过期、可失效）一律 Vue Query**，见 `api-query`；**Pinia 只存客户端状态**（UI 态、草稿、选中项、本地缓存）。
- 唯一豁免：chat 消息缓存留在 Pinia（`features/chat/store/chat.ts`）——流式增量写入的列表是客户端权威状态，头注释已注明拆分模块。新代码不得再造第二个豁免。
- store 不发请求拿"服务端数据"自留（查询走 hooks）；store action 里的请求仅限**命令型写操作**（如 `generate()` 提交生成）且结果回写的是客户端权威状态。

## 结构正典

```ts
/**
 * 图像生成 (生图) 页面 Store (Setup Store).
 *
 * 管理可用模型列表、当前生成参数、当前任务与历史任务数据;
 * 任务状态轮询由 pages/composables/use-task-polling (Vue Query) 驱动,
 * 查询结果经 applyTaskUpdate 回写本 store.
 */
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';

// 注意: 存量 id 为 'imageGen' (camelCase), 已登记反例待改; 新代码一律 kebab-case.
export const useImageGenStore = defineStore('image-gen', () => {
  // ============================================================================
  // State
  // ============================================================================

  /** 当前生成参数. */
  const params = ref<ImageGenParams>(defaultParams());

  /** 是否正在生成. */
  const generating = ref(false);

  // ============================================================================
  // Getters
  // ============================================================================

  /** 当前选中的模型. */
  const selectedModel = computed<ImageGenModelItem | null>(
    () => models.value.find((m) => m.imgModelId === params.value.imgModelId) ?? null,
  );

  // ============================================================================
  // Actions
  // ============================================================================

  /**
   * 发起图像生成 (后台任务).
   *
   * @returns 生成的任务 DTO (pending).
   */
  async function generate(): Promise<ImageGenTask> {
    generating.value = true;
    try {
      const task = await generateImage({ /* ... */ });
      currentTask.value = task;
      return task;
    } catch (err) {
      generating.value = false;
      throw err;
    }
  }

  return {
    params,
    generating,
    selectedModel,
    generate,
    /* ... */
  };
});
```

规则：

1. **分节注释三段固定**：`// ====== State ======`、`// ====== Getters ======`、`// ====== Actions ======`（等号横线与 image-gen store 同款）；段序不得打乱，模块级私有 helper 放 `defineStore` 外部。
2. state 全 `ref<>()`（不用 `reactive`），getter 全 `computed<>()`，action 为普通函数；**每个成员一条中文 JSDoc**。
3. 尾部 `return` 大对象集中导出；不 return 的闭包变量即私有状态。
4. **非响应式私有状态用普通闭包变量**（`let primaryTaskId: string | null = null;`、`const abortControllers = new Map<string, AbortController>()`），不进 ref。
5. 文件头块注释写职责与数据流（谁驱动、数据从哪来回哪去）。
6. 复杂逻辑拆文件：store 目录一文件一职责（`chat/store/` 下 `chat.ts`、`composer-draft.ts`、`context-settings.ts`…），纯函数抽到 `utils/` 并配 spec（见 testing）。
7. 跨 feature 共享的 store 放 `src/shared/<域>/store.ts`（member、providers、update…），不放任何 feature 里。

## 命名

- 导出 `useXxxStore`；文件名 `<域>.ts` 或 `<职责>.ts`（kebab-case）。
- store id **钦点 kebab-case**：`'chat-browser'`、`'collab-transcript'`、`'settings-nav'`；存量 camelCase id 登记反例。
- 目录钦点 `store/`（单数），存量 `stores/` 登记反例。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| options store（`{ state, actions }`） | ~~`collab/store/transcript.ts`~~ **已改写 setup store（2026-08-30，spec 原样通过）** | 历史反例存档 |
| store id 用 camelCase | ~~`'imageGen'`、`'providerModel'`、`'webFetch'`/`'webSearch'`~~ **已整改或随纯查询 store 删除消亡（2026-08-30）** | 历史反例存档；钦点 kebab-case id |
| store 内直接 `useQuery` 存服务端状态 | ~~settings 五域 store~~ **已全部迁移到 hooks 并删除（2026-08-30）** | 服务端状态必须走 Vue Query hooks（api-query）；store 里跑 query 绕开缓存失效体系 |
| 状态目录 `stores/` | ~~全仓~~ **已清零（2026-08-30）** | 钦点 `store/` |
| 持有服务端列表数据不自留 UI 态 | — | 二次豁免禁止，服务端状态归 Vue Query |

## 自查清单

- [ ] 这段状态确实是客户端状态？（服务端状态已回退到 hooks？）
- [ ] setup store + 三段分节注释 + 逐成员中文 JSDoc？
- [ ] id kebab-case、文件在 `store/` 目录、kebab-case 文件名？
- [ ] 非响应式私有状态用闭包变量没进 ref？
- [ ] 没在 store 里 useQuery、没有第二服务端状态缓存？

## 范本文件

- 结构正典：`src/features/image-gen/stores/image-gen.ts`（结构本身为正典；目录名 `stores/` 与 store id `'imageGen'` 是存量反例待改，勿模仿）
- 拆文件多 store：`src/features/chat/store/`（chat.ts / composer-draft.ts / context-settings.ts）
- 共享 store：`src/shared/member/store.ts`、`src/shared/providers/store.ts`

相关技能：`api-query`、`composable`、`feature-organization`、`testing`。
