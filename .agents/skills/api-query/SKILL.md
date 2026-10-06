---
name: api-query
description: 在 tako 前端 对接后端接口、编写 API 封装、Vue Query hooks、错误反馈或轮询逻辑时使用。
---

# API & Vue Query Conventions

服务端状态一律 Vue Query，API 封装一律**函数式**。本文骨架提炼自 `src/features/image-gen/image-gen.api.ts`（API 正典）、`src/features/collab/pages/composables/use-spaces.ts`（mutation 与错误反馈权威）、`src/features/image-gen/pages/composables/use-task-polling.ts`（轮询权威）。

## 适用场景

- 新建 `<域>.api.ts` 或给既有 API 加接口。
- 写 `use-*.ts` 数据 hooks（query/mutation）。
- 实现任务轮询、错误 toast、缓存失效。

## 底座事实（必须先知道）

- 后端**所有响应 HTTP 恒 200**，业务错误靠 `code !== 0` 表达；`src/shared/api/request.ts` 拦截器已把非 0 code 转成 `BusinessError`（携带 code/traceId/details）并自动 `toast.error`（按消息去重）。
- 业务代码拿到的 reject 已是带后端 message 的 Error；要自管错误 UI 时给请求传 `skipErrorToast: true`。
- 通用信封类型在 `src/shared/api/types.ts`：`ApiDataResponse<T>`（单数据）、`ApiListResponse<T>`（列表分页，对应后端 ListResponseSchema）、`PageParams`。

## API 文件正典（函数式 + 解包返回）

```ts
/**
 * 图像生成 (生图) API 模块.
 *
 * 提供生图任务创建、分页查询、删除、可用模型列表.
 */
import request from '@/shared/api/request';
import type { ApiDataResponse, ApiListResponse } from '@/shared/api/types';
import type { ImageGenTask, ImageGenGenerateRequest } from '@/features/image-gen/image-gen.type';

/** API 基础路径. */
const BASE_URL = '/image-gen';

/**
 * 创建图像生成任务.
 *
 * @param payload - 生成请求 (imgModelId, prompt, 可选参数).
 * @returns 任务 DTO.
 */
export async function generateImage(payload: ImageGenGenerateRequest): Promise<ImageGenTask> {
  const response = await request.post<ApiDataResponse<ImageGenTask>>(`${BASE_URL}/generate`, payload);
  return response.data.data;
}
```

规则：

1. 文件头块注释写模块职责；`BASE_URL` 模块级常量；一个后端 controller 一个 `<域>.api.ts`。
2. **函数返回解包后的 DTO，禁止返回信封**——调用方永远 `await getXxx()` 直接拿数据。
3. 每个函数中文 JSDoc + `@param`/`@returns`；按后端路由分组时用 `// ====== 分组名 ======` 分隔注释。
4. 列表接口返回 `{ items, total }` 或直接 `T[]`，由函数内部消化分页信封。

## Vue Query 正典

### queryKey：域常量数组 + as const

```ts
/** 空间列表查询键. */
export const COLLAB_SPACES_QUERY_KEY = ['collab', 'spaces'] as const;

export function useCollabSpaces(includeArchived: MaybeRefOrGetter<boolean> = false) {
  return useQuery({
    queryKey: [...COLLAB_SPACES_QUERY_KEY, computed(() => toValue(includeArchived))],
    queryFn: () => listCollabSpaces(toValue(includeArchived)),
    staleTime: 30_000,
  });
}
```

- 钦点结构：`[域, 资源, ...动态段]`；首个元素是域名字符串，动态段放末尾。
- key 常量导出（UPPER_SNAKE + `as const`），失效逻辑复用常量，禁止散落字面量数组。
- 响应式参数用 `computed(() => toValue(x))` 塞进 key，实现参数变更自动重取。
- 无 key 常量的短命查询（轮询子项）可用字面量 `as const` 数组（`['image-gen', 'task', taskId] as const`）。

### mutation：onError toast 必带，onSuccess 只 invalidate

```ts
/** 提取用户可读错误消息（request 拦截器抛出的 Error 已含后端 message）. */
function errMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) return error.message;
  return fallback;
}

function useInvalidateSpaces() {
  const queryClient = useQueryClient();
  return (spaceId?: string) => {
    void queryClient.invalidateQueries({ queryKey: [...COLLAB_SPACES_QUERY_KEY] });
    if (spaceId) void queryClient.invalidateQueries({ queryKey: [...COLLAB_SPACE_DETAIL_KEY, spaceId] });
  };
}

/** 创建空间. */
export function useCreateCollabSpace() {
  const invalidate = useInvalidateSpaces();
  return useMutation({
    mutationFn: (payload: CollabSpaceCreateRequest) => createCollabSpace(payload),
    onSuccess: (space) => invalidate(space.id),
    onError: (error) => toast.error(errMessage(error, '创建协作空间失败')),
  });
}
```

铁律（源自 AGENTS.md，collab 是权威实现）：

1. **所有 mutation 必带 `onError: (error) => toast.error(errMessage(error, '中文兜底文案'))`，禁止静默失败**——失败无提示在 UI 上的表现就是"点了没反应"。
2. `onSuccess` 只做缓存失效/导航，**不做 toast**（错误才 toast，成功提示仅在确有用户价值时加）。
3. toast 收敛在 hooks 层（`use-*.ts`），调用方只做流程控制（关弹窗/跳转），**不重复 toast**。
4. 同一域的失效逻辑收敛进私有 `useInvalidateXxx()`。
5. **同一提交内多个有依赖的写请求禁止并行编排**：后端有原子接口用原子接口；没有则 `mutateAsync` 串行链。

### 轮询：动态 refetchInterval 指数退避

后台任务型数据（pending → 终态）按 `use-task-polling.ts` 模式：

- 每个待轮询实体一个 query 实例（`useQueries` + computed 动态增减）。
- `staleTime: 0`、`retry: false`、`refetchOnWindowFocus: false`。
- `refetchInterval` 函数：终态返回 `false` 停轮询；进行中 `Math.min(INITIAL * 2 ** count, MAX)` 指数退避（正典 2s→30s 封顶 + 次数上限兜底）。
- 轮询结果经 `watch` 回写 store，数据流规则写在文件头块注释。

## 请求级设置的模式归属（chat-only 红线）

`contextLength` 条数裁剪、`webSearch`/`webFetch` 联网工具挂载等请求级设置**仅对 chat 模式生效**；agent 模式不受这些限制。新增请求级设置先确认模式归属，禁止把 chat-only 开关逻辑复用进 agent 分支。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| class 静态方法风格的 API 文件 | ~~settings 五域 + knowledge 等 9 个~~ **features 内已全部整改（2026-08-30）**；`src/shared/` 下 `update/member/providers` 三个 class API 为存量待改 | 钦点函数式；class 无法 tree-shake |
| 对象字面量风格 API | ~~`chat/mcp-auth.api.ts`~~ **已整改（2026-08-30）** | 历史反例存档 |
| API 函数返回整个信封 | ~~settings 系~~ **已全部整改（2026-08-30）** | 历史反例存档；调用方被迫处理信封 |
| queryKey 用扁平字符串字面量 | ~~settings 五域~~ **已全部整改（2026-08-30，域常量数组 + `shared/query-keys` 单一来源）** | 无域前缀易撞键、无常量可失效 |
| mutation 无 onError toast | ~~`todo updateListMutation`~~ **已补（2026-08-30）**；全量审计已过：其余 19 处定义级无 onError 的 mutation 均由调用方 `mutate(..., { onError })` 覆盖 | 静默失败，违反 AGENTS.md 红线 |

## 自查清单

- [ ] API 函数式 + BASE_URL 常量 + 解包返回 + 中文 JSDoc？
- [ ] queryKey 域常量数组 as const、动态段在末尾、响应式参数经 toValue？
- [ ] 每个 mutation 有 errMessage + toast 兜底、onSuccess 只 invalidate？
- [ ] toast 收敛在 hooks 层、调用方没重复弹？
- [ ] 轮询带退避封顶与终态停轮、结果回写规则写在头注释？

## 范本文件

- API 正典：`src/features/image-gen/image-gen.api.ts`
- mutation/失效/toast 权威：`src/features/collab/pages/composables/use-spaces.ts`
- 轮询权威：`src/features/image-gen/pages/composables/use-task-polling.ts`
- 拦截器与信封：`src/shared/api/request.ts`、`src/shared/api/types.ts`

相关技能：`type-dto`、`pinia-store`、`composable`、`vue-component`。
