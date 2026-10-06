---
name: type-dto
description: 在 tako 前端 定义或修改 DTO 类型（*.type.ts）、组件 props 类型或工具类型别名时使用。
---

# Type & DTO Conventions

DTO 手写、无 codegen；每个 feature 一对 `<域>.type.ts`。本文骨架提炼自 `src/features/image-gen/image-gen.type.ts`（DTO 正典）。

## 适用场景

- 新建/修改 feature 的 DTO 类型文件。
- 决定一段数据用 interface 还是 type、怎么命名。
- 对齐后端 schema 与 JSON 序列化。

## DTO 文件正典

```ts
/**
 * 图像生成 (生图) 相关类型定义.
 *
 * 与后端 schemas/ai/ai_image_gen_schema.py 对应 (驼峰序列化).
 * 生图使用独立的服务商 (ai_image_gen_provider) 与模型 (ai_image_gen_model) 表,
 * 不再复用 chat 模型的 ai_provider / ai_provider_model 表.
 */

/** 生成任务状态. */
export type ImageGenTaskStatus = 'pending' | 'succeeded' | 'failed';

/** 图像生成任务 DTO. */
export interface ImageGenTask {
  /** 任务 ID. */
  taskId: string;
  /** 图像生成模型 ID. */
  imgModelId: string;
  /** 反向提示词. */
  negativePrompt?: string | null;
  /** 参考图 (base64 data URL 或 http(s) 链接). */
  referenceImages?: string[];
}

/** 图像生成请求. */
export interface ImageGenGenerateRequest {
  /** 生成提示词. */
  prompt: string;
  /** 生成数量 (1-4). */
  n?: number;
}
```

规则：

1. **文件头块注释必写后端来源**：对应的后端 schema 文件路径 + 序列化口径（"驼峰序列化"）+ 与 chat 共用表的差异说明。找不到来源的先去 llm-ref 查 schema 再落笔。
2. 字段 **camelCase 直映后端 JSON**，每个字段一条中文 JSDoc；可空写 `?: T | null`（后端可回 null 的字段必须带 `| null`，只有前端可不传的才省略 null）。
3. **对象/DTO/props 用 `interface`；联合字面量、工具别名、常量表推导用 `type`**：

   ```ts
   export type ImageGenTaskStatus = 'pending' | 'succeeded' | 'failed';        // 联合 → type
   export type SseEventType = (typeof SSE_EVENTS)[keyof typeof SSE_EVENTS];    // 推导 → type
   export interface ImageGenTask { /* ... */ }                                  // 对象 → interface
   ```

4. 命名：DTO `XxxTask` / `XxxProvider` 等业务名词；请求体 `XxxCreateRequest` / `XxxUpdateRequest`；**禁 `I` 前缀**。
5. 联合状态字段单独提 `type` 导出（`ImageGenTaskStatus`），运行时判定函数（`isTerminal()`）与该 type 同源，禁止裸字符串比较 `'succeeded'` 散落多处。
6. 通用信封不重复定义：`ApiDataResponse` / `ApiListResponse` / `PageParams` 从 `@/shared/api/types` 引入。
7. 组件内部小类型可放同目录（`chat/pages/components/chat-types.ts`），但会被跨组件引用的必须上提到 `<域>.type.ts`。

## 类型位置

| 内容 | 位置 |
|---|---|
| feature DTO（后端契约） | feature 根 `<域>.type.ts`；子域自带（`settings/mcp/ai-mcp.type.ts`） |
| 通用响应信封 | `src/shared/api/types.ts`（唯一） |
| IPC payload | `src/shared/ipc/contract.ts`（唯一，见 ipc-bridge） |
| 常量表推导 | 常量所在文件内（`SseEventType`） |
| 组件私有小类型 | 同目录或组件内 |

## 与后端对齐的注意点

- 后端 Go/Python schema 是权威；改字段先看后端 schema，**禁止前端单方面"顺手"改字段名**。
- 时间字段后端回字符串（`createdTime?: string | null`），前端不做 Date 反序列化，展示经 `shared/lib/relative-time.ts`。
- 敏感字段标注口径（`api_key 永不回传, 仅 hasApiKey 标记是否已配置`）写进字段 JSDoc。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| `I` 前缀接口 | ~~settings/shared 全域 65 个~~ **已全部整改（2026-08-30）** | 历史反例存档；新代码出现即打回 |
| 请求体 `Req`/`Body` 混杂后缀 | ~~settings Req 系 + todo/flash-note `*Body` 家族~~ **已全部统一 `*Request`（2026-08-30）** | 历史反例存档 |
| 文件头缺后端 schema 来源 | — | DTO 漂移无从核对 |
| 对象字面量用 `type` / 联合用 `interface` | — | 见 interface vs type 规则 |

## 自查清单

- [ ] 文件头写了后端 schema 来源与序列化口径？
- [ ] 字段 camelCase + 逐字段中文 JSDoc + 可空 `?: T | null`？
- [ ] interface/type 对号入座、无 I 前缀、Request/Response 后缀？
- [ ] 状态联合提了 type、没有裸字符串比较散落？
- [ ] 信封类型从 shared/api/types 引、没重复定义？

## 范本文件

- DTO 正典：`src/features/image-gen/image-gen.type.ts`
- 通用信封：`src/shared/api/types.ts`
- IPC 类型：`src/shared/ipc/contract.ts`

相关技能：`api-query`、`code-style-guide`、`ipc-bridge`、`sse-streaming`。
