# Tako 前端开发规范（技能索引）

本套技能覆盖 tako 前端（Wails 3 webview + Vue 3 + TS + Tailwind v4）的**代码写法细节层**，目标只有一个：**任何模型/任何人照本套规范写代码，产出风格一致**。实现手段：

1. **双轨必裁决**——多种写法并存处全部钦点唯一正典，旧写法登记反例表，新代码禁止。
2. **正典骨架可照抄**——写新代码先找对应技能的「范本文件」读一遍再落笔。
3. **反例指向真实文件**——长得像反例的存量代码是要改的债，不是可以模仿的先例。
4. **自查清单收尾**——每个任务完成前对照清单自检。

> 迁移注记：本套技能自上游前端规范迁移改写。feature-organization / code-style-guide / vue-component / api-query / pinia-store / composable / type-dto / testing 八份的正文方法论照旧生效；遇到文中残留的上游项目语境（路径、命令、组件名）时，以 tako 实际目录与命令为准。

## 技能索引

| 技能 | 管什么 |
|---|---|
| [feature-organization](feature-organization/SKILL.md) | feature/子域目录结构、文件归属、命名 |
| [code-style-guide](code-style-guide/SKILL.md) | 命名总表、import 排序、注释风格、eslint 禁区速查 |
| [vue-component](vue-component/SKILL.md) | SFC 结构、props/emits、registry 组件消费、交互反馈 |
| [api-query](api-query/SKILL.md) | API/bridge 函数式封装、mutation 错误反馈 |
| [pinia-store](pinia-store/SKILL.md) | setup store 正典、服务端/客户端状态分工 |
| [composable](composable/SKILL.md) | 数据/副作用 hooks、createGlobalState 单例 |
| [type-dto](type-dto/SKILL.md) | DTO 文件、interface vs type、契约对齐 |
| [bridge](bridge/SKILL.md) | wails 桥接三同步（contracts→bindings→adapter） |
| [provider-streaming](provider-streaming/SKILL.md) | 翻译流式事件消费、多开子流聚合 |
| [testing](testing/SKILL.md) | vitest 纯函数必测、spec 正典、不测范围 |

## 先看这个

**"仓库里存在两种写法" ≠ "两种都对"。** 遇到存量代码与本文冲突时，以本文钦点正典为准；批量整改登记 TODO，禁止新代码跟随存量旧风格。

## 与既有文档的分工

| 文档 | 管什么 |
|---|---|
| `AGENTS.md`（仓库根） | 全局元规则与高频红线 |
| `docs/spec/` | SPEC-A（Go 后端）/ SPEC-B（前端）分工规格 |
| `packages/contracts/README.md` | 桥接契约、流式事件时序、SQLite schema |
| **本套技能** | 上述文档不管的"代码怎么写"细节层 |

## 场景速查（要做 X → 读哪份）

| 场景 | 读 |
|---|---|
| 新建一个页面/feature | `feature-organization` → `vue-component` |
| 调一个 Go 能力 | `bridge`（contracts 对齐，禁直接 import bindings） |
| 消费翻译流式事件 | `provider-streaming` |
| 新建客户端状态 | `pinia-store`（先确认不是服务端状态） |
| 抽一段复用逻辑 | `composable` |
| 写了个纯函数 | `testing` |
| 任何文件 | `code-style-guide`（命名/import/注释/禁区） |

## 验证命令（按改动范围收窄使用）

```bash
cd apps/desktop/frontend
bun run lint        # eslint（boundaries + token 护栏 + prettier）
bun run test        # vitest
bun run typecheck   # vue-tsc -b
bun run build       # vue-tsc -b + vite build
```

依赖一律 Bun（`bun install`），禁 npm/pnpm/yarn。仓库根 `make ci` 会先生成 bindings 再全量检查。

## 一句话总结

**先读对应技能的范本文件，照正典骨架写；所有"都行"处按钦点裁决；写完过自查清单。**
