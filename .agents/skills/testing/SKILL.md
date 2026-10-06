---
name: testing
description: 在 tako 前端 编写或修改 vitest 测试、为新纯函数补 spec 或变更既有行为时使用。
---

# Testing Conventions

vitest 4（environment: node），spec 与被测文件**同目录同名**。本文骨架提炼自 `src/features/chat/store/context-settings.spec.ts`（正典）。

## 适用场景

- 新建纯函数后补 spec。
- 变更既有纯函数行为（先补测试锁定旧语义，再改）。
- 评审测试结构与命名。

## 铁律

1. **纯函数层必须有测试**：`stream/codec.ts`、`store/display-messages.ts`、`shared/lib/*`、feature `utils/*`、常量归一化函数——凡是被 sse-streaming / display-messages 链路消费的纯函数，改它必配 spec。
2. spec 位置：被测文件旁，`<名>.spec.ts`；vitest include 已覆盖 `src/**/*.spec.ts` 与 `electron/**/*.spec.ts`，无需登记。
3. **组件渲染不测**（现状无组件测试链），组件逻辑抽成纯函数/composable 后测函数本身。
4. 行为变更前先补测试锁定当前语义（失败→改函数→绿），禁止先改后补到"能过就行"。
5. 验证命令：`bun run test`（全量）/ `bun run test:watch`；校验收窄原则下，改哪测哪。

## spec 正典骨架

```ts
import { describe, expect, it } from 'vitest';

import { normalizeConversationContextSettings } from './context-settings';

describe('normalizeConversationContextSettings', () => {
  it('补齐历史会话缺失的网络工具配置并默认关闭', () => {
    const result = normalizeConversationContextSettings({
      contextLength: 20,
      reasoningEnabled: true,
    });

    expect(result).toMatchObject({
      contextLength: 20,
      reasoningEnabled: true,
      webSearch: { isEnabled: false },
    });
  });

  it('推理档位标记兼容 snake_case 历史键', () => {
    const result = normalizeConversationContextSettings({
      contextLength: 20,
      reasoning_level_set: true,
    } as unknown as Parameters<typeof normalizeConversationContextSettings>[0]);
    expect(result.reasoningLevelSet).toBe(true);
  });
});
```

规则：

1. **显式 import**（`import { describe, expect, it } from 'vitest'`），不依赖 globals 隐式注入。
2. `describe` 用**被测函数名**；`it` 用**中文行为描述**（说清"输入什么 → 该怎样"）。
3. 断言风格：部分匹配用 `toMatchObject`，精确相等用 `toEqual`/`toBe`；一个 it 只测一个行为。
4. 历史脏数据兼容断言用 `as unknown as Parameters<typeof fn>[0]` 显式声明"这是超出当前类型的输入"。
5. 时间相关测试用真实日期构造（`new Date('2026-08-15T12:00:00')`）后冻结/注入，不 mock 系统时钟。
6. 异步/流式回调测试直接构造事件对象数组喂纯函数，不起真实 SSE 连接。

## 不测范围

| 内容 | 处理 |
|---|---|
| .vue 组件渲染 | 不测；逻辑抽纯函数后测函数 |
| electron main 生命周期 | 纯逻辑部分（如 `backend-process-registry.ts`、`collab-workspace.ts`）照常 spec |
| `shared/api/request.ts`（axios 实例副作用） | 不测；其类型与信封由消费方 spec 间接覆盖 |

## 已知反例（存量发现即登记 TODO，新代码禁止）

| 反例 | 位置 | 为什么错 |
|---|---|---|
| 新增/修改纯函数不配 spec | — | 纯函数层必测铁律；spec 是重构唯一安全网 |
| spec 放独立 `__tests__/` 或 `tests/` 目录 | — | 钦点同目录同名，搬迁成本高 |
| describe 用中文、it 用英文或反之 | — | 正典：describe 函数名（英文）、it 行为（中文） |
| mock 一切（vi.mock 铺满）而不是构造输入 | — | 纯函数应直接喂对象断言输出 |

## 自查清单

- [ ] 改的每个纯函数都有同目录同名 spec？
- [ ] describe=函数名、it=中文行为、显式 import vitest？
- [ ] 行为变更先锁旧语义再动函数？
- [ ] 没有 vi.mock 铺满、组件渲染强行测？
- [ ] `bun run test` 过且没引入新全局失败？

## 范本文件

- spec 正典：`src/features/chat/store/context-settings.spec.ts`
- 流式 codec spec：`src/features/chat/stream/codec.spec.ts`
- 时间工具 spec：`src/shared/lib/relative-time.spec.ts`
- electron 侧 spec：`electron/backend-process-registry.spec.ts`

相关技能：`sse-streaming`、`api-query`、`composable`、`code-style-guide`。
