# SPEC-A · Go 后端（执行者：swe2 · 分支 feat/backend）

> 开工前必读：仓库根 `AGENTS.md`、`packages/contracts/README.md`（契约与时序）、`apps/desktop/internal/provider/types.go`（**接口已定稿，不得改签名**）。
> 参考物（不照抄，按需裁剪）：`~/llm-ref/internal/infra/ai/llm/`（先读其 README）。

## 边界与纪律

- 只写 `apps/desktop/internal/**`（含 `*_test.go`）；bridge 的实装体可填，但注册结构与方法签名不动
- contracts 只读；需要变更走 PR「契约变更申请」
- 提交前：`go vet ./... && go test ./...` 全绿；涉及 bindings 时通知主线跑 `make generate`

## 任务与顺序

### A1 LLM provider（最先做，联调价值最大）

`internal/provider/` 下新建：

| 文件 | 内容 |
|---|---|
| `errors.go` | `ErrorKind`（types.go 已定义 8 个 ErrorCode）→ 对应 HTTP/响应错误的分类函数 `Classify(err) ErrorCode`；瞬时错误判定（429/connect 可重试） |
| `httpx.go` | 共享流式 `http.Client`：`Timeout=0` + `ResponseHeaderTimeout`（首包 300s）+ 空闲看门狗 ReadCloser（连续 300s 无数据断流）。参考 llm-ref `http_client.go`（~150 行，可近乎原样借鉴） |
| `retry.go` | `StreamWithRetry(ctx, p Provider, req, emit, max=2)`：**零产出才重试**（已 emit delta 的流不重试，防重复文本）+ 指数退避 + emit `stream-retry` 事件 |
| `llm/openai.go` | OpenAI 兼容适配器：`POST {baseUrl}/chat/completions`，SSE 主循环（`data: ` 行 + `[DONE]` 终态），**截断防御**：干净 EOF 未见终态标记 → 合成断流错误交给重试层。鉴权按 `authScheme`（bearer=Authorization: Bearer；x-api-key=x-api-key 头）；`customHeaders` 先写入、鉴权头最后覆盖（防伪造）。系统提示词模板：翻译到目标语言，只回译文 |

验收：`httptest` 流式 mock 覆盖 ①正常流 ②中途 error 行 ③无终态标记断流→重试→成功 ④401→auth ⑤零产出超时重试。

### A2 传统翻译三件套

`internal/provider/{deel,google,microsoft}.go`，各实现 `Provider` 接口：普通 HTTP 请求，成功后**一次性 emit delta + finish**（时序见 contracts README）。

- `deel`：DeepL Free API（api-free.deepl.com，`auth_key` 头）；语言映射 zh-Hans→ZH
- `google`：网页免费接口（translate.google.com/translate_a/single，tk 参数可省的路由）；zh-Hans→zh-CN
- `microsoft`：Edge 免费接口（api-edge.cognitive.microsofttranslator.com 或等价免费路由）；无需 key 优先

验收：httptest mock 参数组装 + 响应解析；真实 key 冒烟做成 `//go:build smoke` tag 的可选测试。

### A3 store（SQLite）

`internal/store/`：`modernc.org/sqlite`；migrate 读 `packages/contracts/src/schema.sql`（embed 进 Go 或复制常量，**以 contracts 为唯一依据**）；库路径 `os.UserConfigDir()/Tako/tako.db`。

CRUD：Services（List 按 order 升序 / Save / Delete）、History（分页 List 含 keyword/serviceId/favoriteOnly 过滤 + total / Save / Delete / ToggleFavorite / Clear）、Settings（KV JSON）。
ID 生成用 `nanoid`（`github.com/matoous/go-nanoid/v2`）或等价。

验收：`t.TempDir()` 库的 CRUD 测试 + 迁移幂等（跑两次 migrate 无错）。

### A4 取词（最难，预留最多时间）

`internal/clipboardx/`：
- macOS：NSPasteboard（cgo，`#cgo LDFLAGS -framework AppKit`）读文本/写文本/保存全部类型再恢复
- Windows：syscall OpenClipboard/GetClipboardData（CF_UNICODETEXT），保存恢复同理
- 竞态窗口 <100ms：保存 → 写入 → 等新值 → 恢复，全链路超时保护

`internal/selection/`：
- macOS：CGEvent 模拟 ⌘C（需辅助功能权限，`AXIsProcessTrustedWithOptions` 检测）+ 权限检测/引导函数（bridge.CheckAccessibility/OpenAccessibilitySettings 实装时接这里）
- Windows：SendInput 模拟 Ctrl+C
- 流程：快捷键触发 → 模拟复制 → 轮询剪贴板变化（300ms 超时）→ 拿到文本 → 恢复剪贴板 → 返回文本

验收：可单测部分（状态机/超时/恢复序列）用 go test；平台行为做 `cmd/selection-smoke` 手动冒烟（`go run ./cmd/selection-smoke`）。

### A5 bridge 实装

`internal/bridge/service.go` 的 stub 换成真实现。**依赖装配走生命周期钩子，`NewService()` 保持无参、main.go 零改动**（main.go 是主线独占，feat/backend 分支不触碰）：

- `ServiceStartup(ctx, options)`（钩子签名已留在 stub）：打开 store（`os.UserConfigDir()/Tako/tako.db`）→ 建 provider 注册表（map[type]Provider 构造器）→ 初始化 `active map[requestID]context.CancelFunc`。装配失败返回 error 让应用启动即败露
- `ServiceShutdown()`：关 store、cancel 全部在途请求
- `Translate`：生成 requestId → 存 cancelFunc → 每服务 goroutine fan-out → emit 用 `application.Get().Event.Emit("translate:event", ev)`（`application.Get()` 是全局单例，ServiceStartup 后即可用；事件名与 payload 镜像 contracts 的 `BridgeEvents`/`StreamEvent`）；全部终态后清 map 条目
- `CancelTranslate`：调 cancelFunc（provider ctx 取消 → 该服务 emit error(code=cancelled)）
- 事件结构体直接用 `provider.StreamEvent`（字段已对齐 contracts）

若确实需要 main.go 配合（例如未来加托盘/快捷键服务注册），在 PR 描述里列「main.go 变更申请」，主线代改——不要在 feat/backend 分支提交 main.go。

验收：`go build ./...` 通过 + 生命周期/store 单测；`wails3 dev`（主线环境）前端真实调用全链路；与主线约 I1 联调。

## 时序与依赖

```
A1 ──> A2 ∥ A3 ──> A4 ──> A5
       （A5 可随各模块完成逐步实装，不必等全部）
```

B（前端）在 mock 上并行开发，你无需等待；I1 联调只依赖 A5+B3 两个最早完成点。

## 参考文件精读清单（llm-ref llm，按优先级）

1. `README.md`（装配链与不变量速查）
2. `http_client.go`（看门狗+超时模型 → A1 httpx.go）
3. `openai_compatible.go` + `openai_chat.go`（SSE 主循环 + chunk 解析 → A1 llm/openai.go）
4. `errors.go`（ErrorKind 分类 → A1 errors.go）
5. `retry.go`（零产出才重试 → A1 retry.go）

明确不抄：Anthropic 适配器、Responses 子格式、工具调用、reasoning variant、ProviderCache、ConfigResolver、多模态。
