# @tako/contracts

前后端桥接契约单一来源。Go 壳与 Vue 前端共享此处的类型、方法名、事件名与 SQLite schema。

- `src/index.ts` — 全部 TS 类型 + `BridgeMethods` / `BridgeEvents` 常量 + `Bridge` 接口 + 语言表
- `src/schema.sql` — SQLite schema v1（`internal/store` 迁移依据）

## Go 侧映射规则

- `internal/bridge` 中的服务方法使用 `BridgeMethods` 中的**导出方法名**（PascalCase）
- 参数/返回值 struct 的 json tag 与 TS 字段**逐字段 camelCase 对齐**，例如 `ServiceSetting` ↔ `type ServiceSetting struct { BaseURL string `json:"baseUrl,omitempty"` ... }`
- `StreamEvent` 由 Go 构造后经 `application.Get().EmitEvent(BridgeEvents.translateEvent, event)` 推送
- Wails3 自动生成 TS bindings，前端**不得直接 import bindings**，一律经 `src/shared/desktop/adapter.ts`

## 流式事件时序（translate）

一次 `translate(req)`（`serviceIds` 可多个，多开 fan-out），每个 serviceId 一条独立子流：

```
message-start → delta* → (stream-retry)* → finish          # 正常
message-start → (stream-retry)* → error                    # 失败
message-start → delta* → error(errorCode=cancelled)        # 前端 cancelTranslate
```

约定：

- **message-start**：该服务连接建立、即将产出。收到前 UI 显示 loading
- **delta**：增量文本。LLM 逐 token；传统服务（deepl/google/microsoft）请求完成后**一次性**发单个 delta
- **stream-retry**：零产出断流时的自动重试（attempt 从 1 递增），UI 可提示"重试中"
- **finish**：正常终态，text 为完整译文（usage 可选）
- **error**：终态错误。errorCode ∈ auth / rate_limit / timeout / connect / context_overflow / network / cancelled / unknown
- 每个事件都携带 `requestId` + `serviceId`，前端按 `(requestId, serviceId)` 聚合
- `finish`/`error` 后该子流不再有事件；全部子流终态后本次 requestId 结束

## 语言代码

BCP-47 简化（`zh-Hans`、`en`、`ja`…），全表见 `index.ts` 的 `LANGUAGES`。Go 侧负责映射到各服务自有代码（如 DeepL `ZH`、Google `zh-CN`）。

## 变更流程

本包仅主线可改。需要变更：在分支 PR 描述中列明「契约变更申请」（新增/修改的字段与理由），主线合并后 `make generate` 重新生成 bindings。
