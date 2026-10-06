---
name: bridge
description: 新增/修改前后端桥接能力（Go 方法、contracts 类型、adapter）时使用。契约三同步流程与验证命令。
---

# bridge：wails 桥接三同步

改桥接能力必须三处同步，缺一不可：

```
packages/contracts/src/index.ts   ← 契约唯一真源（类型 + BridgeMethods + BridgeEvents）
       ↓ make generate
apps/desktop/internal/bridge/     ← Go 服务方法（json tag 对齐 contracts camelCase）
       ↓ wails3 自动生成
frontend/bindings/                ← TS bindings（生成物，gitignore，禁止手改）
       ↓
frontend/src/shared/desktop/adapter.ts  ← 业务唯一入口（wails/mock 双模式）
```

## 铁律

1. **先改 contracts，再改 Go，最后 adapter**。contracts 是唯一真源；分支上需要契约变更时在 PR 描述提「契约变更申请」，主线合并，禁止在 feat 分支直接改 contracts。
2. **业务代码禁止 import `frontend/bindings/**` 与 `@wailsio/runtime`**（eslint no-restricted-imports 强制），一律 `const bridge = await desktop()`。adapter 是 wails 的唯一触点。
3. Go 方法名 = `BridgeMethods` 的 PascalCase；参数/返回值 struct 的 json tag 与 TS 字段逐字段对齐（如 `BaseURL string \`json:"baseUrl,omitempty"\``）。
4. Go→前端推送一律走 `BridgeEvents` 常量里的事件名，payload 结构在 contracts 定义；adapter 侧加对应 `onXxx` 订阅方法。
5. mock 实现（`src/shared/desktop/mock.ts`）必须与真实行为同构——每个新方法/事件在 mock 里同步补齐，让前端离开 Go 也能全流程开发。

## 流程（新增一个能力）

1. contracts：加类型（需要时）+ `BridgeMethods` 加方法名 + `Bridge` 接口加签名
2. Go：`internal/bridge/types.go` 加镜像 struct，`service.go` 加导出方法
3. `make generate` 重新生成 bindings（11s 左右，幂等）
4. adapter：wails 分支转发 + mock 实现补齐
5. 验证（按序，全绿才算完）：

```bash
cd apps/desktop && go build ./...        # Go 侧编译
make generate && cd frontend && bun run typecheck && bun run test
```

## 范本文件

- contracts：`packages/contracts/src/index.ts`（全部现有类型的写法基准）
- Go 镜像：`apps/desktop/internal/bridge/types.go`
- adapter：`apps/desktop/frontend/src/shared/desktop/adapter.ts`（动态 import 双模式）
- mock：`apps/desktop/frontend/src/shared/desktop/mock.ts`

## 自查清单

- [ ] contracts 先行修改，分支未直接动 contracts
- [ ] Go json tag 与 TS 字段逐字段对齐（camelCase）
- [ ] 事件名走 BridgeEvents 常量
- [ ] mock 同步实现，vitest 覆盖新能力
- [ ] `go build` + `typecheck` + `test` 三绿
