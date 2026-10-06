---
name: provider-streaming
description: 消费翻译流式事件（translate:event 的 StreamEvent）、实现多开对照渲染时使用。子流聚合、渐进渲染、终态处理。
---

# provider-streaming：翻译流式事件消费

契约见 `packages/contracts/README.md`「流式事件时序」。要点：

```
一次 translate() = requestId
每个 serviceId 一条独立子流：
  message-start → delta* → (stream-retry)* → finish | error
```

## 铁律

1. **按 `(requestId, serviceId)` 聚合子流**。一次多开翻译有 N 条并发子流，事件交错的，禁止假设事件按服务分组到达。
2. **传统服务只有一次性 delta**（deepl/google/microsoft 完成后单个 delta + finish），LLM 才有逐 token 流。渲染层必须同时兼容两种形态——不要给 delta 间隔做任何假设。
3. **finish/error 是终态**，该子流不再有事件；全部子流终态后本次请求结束（用于"翻译中"全局态的归零）。
4. **stream-retry 不是错误**：零产出自动重试（attempt 递增），UI 提示"重试中"即可，禁止把重试渲染成失败。
5. **error 有 errorCode**（auth/rate_limit/timeout/connect/context_overflow/network/cancelled/unknown），提示文案按 code 分档：auth → 引导检查密钥，rate_limit → 提示稍后再试，其余给 message。
6. **取消**：用户取消调 `cancelTranslate(requestId)`，收到 `error(code=cancelled)` 后清理 UI 态（不算失败提示）。

## 渲染正典（B3 参照实现）

- store 内维护 `Map<serviceId, StreamState>`：`status: 'pending' | 'streaming' | 'retrying' | 'done' | 'error'`、`text`（delta 累积）、`error`
- delta 到达：`state.text += e.text`，`status = 'streaming'`；finish 到达：以 `e.text` 为准覆盖累积值（防丢字）
- 订阅生命周期：进入页面 `bridge.onTranslateEvent(handler)`，离开时调用返回的取消函数——**禁止在组件内多处重复订阅**

## 范本文件

- mock 流（测试基准）：`apps/desktop/frontend/src/shared/desktop/mock.ts` 的 `translate`
- 消费示例：`apps/desktop/frontend/src/shared/desktop/adapter.spec.ts` 的「start→delta*→finish」用例

## 自查清单

- [ ] 聚合键为 (requestId, serviceId)
- [ ] 兼容单 delta（传统服务）与多 delta（LLM）两种形态
- [ ] finish 覆盖而非追加
- [ ] stream-retry / cancelled 不渲染为失败
- [ ] 组件卸载取消订阅
