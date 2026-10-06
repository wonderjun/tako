---
name: code-style-guide
description: 在 tako 前端 编写或修改任何 TS/Vue 文件时使用（命名、import 排序、注释风格与 eslint 禁区速查）。
---

# Code Style Guide

格式由 Prettier 统一（`printWidth: 120`、`semi: true`、`singleQuote: true`、`trailingComma: "all"`、`arrowParens: "always"`、`endOfLine: "lf"`，经 eslint-plugin-prettier 强制），**不要手工对齐格式**。本文只管 Prettier 管不了的：命名、结构顺序、注释与禁区。范例见 `src/features/image-gen/image-gen.api.ts`（注释与命名正典）。

## 适用场景

- 新建任何 `.ts` / `.vue` 文件。
- 评审或修改现有代码的命名与注释。
- 拿不准某个写法会不会撞 eslint 禁区。

## 命名总表

| 对象 | 规则 | 正例 |
|---|---|---|
| 文件（非组件 .ts） | kebab-case | `use-task-polling.ts`、`chat-message-normalize.ts` |
| 文件（业务 .vue） | kebab-case | `result-gallery.vue` |
| 文件（API/类型对） | `<域>.api.ts` / `<域>.type.ts` | `image-gen.api.ts` |
| 变量/函数 | camelCase，动词开头 | `loadHistory()`、`buildImageGenMediaUrl()` |
| 常量（模块级） | UPPER_SNAKE_CASE | `MAX_POLL_INTERVAL_MS`、`SSE_EVENTS` |
| 类型/接口 | PascalCase，**无 I 前缀** | `ImageGenTask`、`CollabSpaceCreateRequest` |
| 请求/响应类型 | `XxxCreateRequest` / `XxxUpdateRequest` 后缀 | `ImageGenProviderCreateRequest` |
| 类型别名（联合/推导） | PascalCase | `ImageGenTaskStatus`、`SseEventType` |
| Pinia store | `useXxxStore` + kebab-case id | `defineStore('chat-browser', ...)` |
| composable | `use-xxx.ts` 文件 + `useXxx` 导出 | `useImageGenTaskPolling` |
| queryKey 常量 | UPPER_SNAKE + `as const` | `COLLAB_SPACES_QUERY_KEY` |
| 布尔 | `is/has/can/should` 前缀 | `isLoadingModels`、`isTerminal()` |

## import 排序

固定五段，段内按路径字母序，段间空行（prettier 不排 import，必须手排）：

```ts
// ① vue / vueuse 核心
import { computed, ref, watch } from 'vue';
import type { MaybeRefOrGetter } from 'vue';
// ② 第三方（vue-query、vue-sonner、lucide-vue-next、pinia…）
import { useMutation, useQuery } from '@tanstack/vue-query';
import { toast } from 'vue-sonner';
// ③ registry 组件库
import { Button } from '@/components/ui/button';
// ④ feature 内（@/features/...）与 shared（@/shared/...）
import { getImageGenTask } from '@/features/image-gen/image-gen.api';
import request from '@/shared/api/request';
// ⑤ 相对路径（同目录子组件、spec）
import ImageGenerationLoader from './image-generation-loader.vue';
```

`import type` 与值 import 分行写（`import type { ImageGenTask } from ...`）。

## 注释风格

- **全部中文**，块注释用 JSDoc（`/** */`），句末带英文句号（`.`），与存量一致。
- 每个导出函数/类型/字段一条 JSDoc；API 函数带 `@param` / `@returns`。
- 文件头块注释写清**模块职责与数据流**（谁驱动谁、数据往哪流），参考 `use-task-polling.ts` 的 15 行头注释。
- 注释只写"为什么"和契约约束（"与 settings/about-update 同源"、"前后端契约镜像，双端改动必须同步"），不复述代码。
- 行内注释用 `//`，前置空行隔开。

```ts
/** API 基础路径. */
const BASE_URL = '/image-gen';

/**
 * 拼接生图媒体文件的完整 URL (供 <img> 直接引用).
 *
 * 与 axios 实例同款 baseURL 规则: 开发模式走 Vite proxy, 生产模式直连后端.
 */
export function buildImageGenMediaUrl(path: string): string {
```

## 结构顺序（script setup / ts 文件）

`.ts` 文件：文件头块注释 → imports → 模块级常量 → 私有 helper → 导出。
`.vue` 文件与 SFC 细则见 `vue-component`。

## eslint 禁区速查（error 级，撞上即 CI 失败）

| 禁区 | 规则来源 | 正确做法 |
|---|---|---|
| renderer 任何地方摸 `window.desktop` | no-restricted-syntax | 走 `src/shared/desktop/adapter.ts` 能力函数 |
| renderer import `electron` | no-restricted-imports | 经 IPC 契约与 adapter（见 ipc-bridge） |
| features 互引、features→app/shell | boundaries | 上移 shared/ |
| 裸 Tailwind 色板（`bg-red-500` 等 22 色系）、任意 hex `[#...]`、`dark:bg-white` 双写、任意 `text-[Npx]` | vue/no-restricted-class（features+shell 的 .vue） | 用语义 token（`bg-primary`、`text-muted-foreground`、`text-body`），视觉契约见 `tako 前端/DESIGN.md` |
| `no-explicit-any` | tseslint recommended | 具体类型；历史脏数据用 `as unknown as Parameters<typeof fn>[0]` 收敛 |

registry 目录（`components/ui/`、`ai-elements/`、`tiptap-editor/`）豁免部分类型规则，业务代码不得效仿。

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| 类型带 `I` 前缀 | ~~settings/shared 全域 13 个文件~~ **已全部整改（2026-08-30）** | 历史反例存档；同一仓库两套类型命名 |
| 注释缺 JSDoc 或用英文散文 | — | 与全仓中文 JSDoc 风格割裂 |

## 自查清单

- [ ] import 五段排好、`import type` 分行？
- [ ] 命名对照总表（尤其类型无 I 前缀、布尔 is/has 前缀）？
- [ ] 每个导出有中文 JSDoc、文件头有职责与数据流说明？
- [ ] 没有裸色板 / hex / dark: 双写 / 任意像素值？
- [ ] 没摸 `window.desktop`、没 import electron？

## 范本文件

- 注释与命名：`src/features/image-gen/image-gen.api.ts`、`src/features/image-gen/image-gen.type.ts`
- 头注释数据流：`src/features/image-gen/pages/composables/use-task-polling.ts`
- 禁区配置：`tako 前端/eslint.config.js`、`tako 前端/.prettierrc`

相关技能：`vue-component`、`type-dto`、`feature-organization`、`ipc-bridge`。
