# Tako AGENTS.md

> 多触手聚合翻译（Wails 3 + Go + Vue 3）。对标 Bob 的核心体验，资源 100% 原创。

## 文档导读

| 要做什么 | 读哪 |
|---|---|
| 写前端代码（任何 .vue/.ts） | `.agents/skills/dev-conventions.md`（索引）→ 对应技能 |
| 写 Go 后端（internal/**） | `docs/spec/SPEC-A.md` |
| 动桥接（contracts/bindings/adapter） | `.agents/skills/bridge/SKILL.md` |
| 消费翻译流式事件 | `.agents/skills/provider-streaming/SKILL.md` |
| 查类型/事件/schema 契约 | `packages/contracts/`（README + index.ts + schema.sql） |
| 提交代码 | `upstream` 系 git-workflow 惯例：conventional commits |

## 技术栈

- **桌面壳**：Wails 3 v3.0.0-beta.28（锁版本），Go ≥ 1.27
- **前端**：Vue 3.5 + Vite 8 + TS 6 + Tailwind v4 + shadcn-vue（reka-ui）+ Pinia
- **存储**：modernc.org/sqlite（纯 Go）
- **工具链**：Bun（前端）、Makefile（编排）、Vitest + go test

## 高频命令（仓库根）

```bash
make dev          # wails3 dev 热重载（先 bun install）
make ci           # generate bindings + 前端 lint/typecheck/test + go vet/test
make build        # 当前平台生产包
make generate     # bridge 契约变更后重新生成 bindings
```

前端目录内：`bun run lint / typecheck / test / build`。

## 目录地图

```
apps/desktop/
├── main.go               # 窗口注册与服务装配（主线维护）
├── internal/             # Go 模块（SPEC-A 领地）
│   ├── provider/         # 翻译适配器：types.go 定稿的 Provider 接口
│   ├── bridge/           # wails 服务层：方法名对齐 contracts BridgeMethods
│   ├── store/            # SQLite（schema 见 packages/contracts/schema.sql）
│   ├── selection/        # 取词（A4）
│   ├── clipboardx/       # 剪贴板保存恢复（A4）
│   ├── hotkey/           # 全局快捷键
│   └── license/          # 许可证桩（商业化解锁点）
└── frontend/             # Vue（SPEC-B 领地）
    ├── src/app/          # 入口/路由/页面
    ├── src/components/ui/# shadcn-vue（registry 复制式，不手工维护）
    ├── src/features/     # feature-first 业务域
    └── src/shared/desktop/  # adapter.ts（wails 唯一触点）+ mock
packages/contracts/       # 桥接契约唯一真源（仅主线可改）
docs/spec/                # SPEC-A / SPEC-B 分工规格
```

## 协作红线（并行分支铁律）

1. **分支**：`main`（主线）/ `feat/backend`（swe2）/ `feat/frontend`（flash）
2. **文件边界**：
   - feat/backend 只写 `apps/desktop/internal/**`
   - feat/frontend 只写 `frontend/src/{app,components,features}/**`
   - 主线独占：根配置、`packages/contracts/**`、`main.go`、`internal/bridge/`（实装体可填，注册结构不动）、`frontend/src/shared/desktop/adapter.ts`
3. **contracts 变更申请制**：分支需要契约变更时在 PR 描述列明字段与理由，主线统一改。
4. bindings 是生成物（gitignore）：任何 lint/typecheck/build 前保证跑过 `make generate`。

## 高频红线

- 前端禁止直接 import `bindings/**` 与 `@wailsio/runtime`（eslint 强制），一律 `await desktop()`
- Go 侧 json tag 必须与 contracts camelCase 逐字段对齐
- `internal/provider/types.go` 的接口签名已定稿，实现方不得改签名
- UI 硬编码颜色禁止（eslint token 护栏），色值走 `src/app/style.css` token
- 依赖一律 Bun，禁 npm/pnpm/yarn
- 不复制 Bob 的任何素材（UI/文案/图标），功能对齐、资源原创
