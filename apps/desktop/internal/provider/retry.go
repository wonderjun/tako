// 流式调用的自动重连：核心不变量是「零产出才重试」——tracked emit 见到
// 任何非 stream-retry 事件即标记产出，已产出内容的失败原样上抛（防正文
// 重复流出）。参考 llm-ref retry.go 裁剪（去掉 RetryExhaustedError，
// 翻译场景重试预算小，上抛原始错误即可）。
package provider

import (
	"context"
	"time"
)

// 重连退避：base 倍增、cap 封顶。包级变量便于测试压缩等待（改写后需恢复）。
var (
	streamRetryBase = 500 * time.Millisecond
	streamRetryCap  = 4 * time.Second
)

// streamRetryBackoff 第 attempt 次重连（1 起）前的退避时长。
func streamRetryBackoff(attempt int) time.Duration {
	backoff := streamRetryBase
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if backoff >= streamRetryCap {
			return streamRetryCap
		}
	}
	if backoff > streamRetryCap {
		return streamRetryCap
	}
	return backoff
}

// StreamWithRetry 包装一次流式翻译调用：遇瞬时错误（IsTransient）且本次
// 调用尚未产出任何用户可见事件时，先 emit stream-retry 事件，再按指数
// 退避重试同一调用。
//
// 约定（对齐 contracts 时序）：stream-retry 事件的 attempt 从 1 递增；
// emit 失败或 ctx 取消立即返回；maxRetries 为首次失败后的最大重连次数
// （总尝试 = maxRetries + 1），<=0 表示不重试。
func StreamWithRetry(
	ctx context.Context,
	p Provider,
	req TextReq,
	emit func(*StreamEvent) error,
	maxRetries int,
) error {
	if maxRetries < 0 {
		maxRetries = 0
	}
	for attempt := 0; ; attempt++ {
		produced := false
		trackedEmit := func(ev *StreamEvent) error {
			if ev != nil && ev.Type != EventStreamRetry {
				produced = true
			}
			return emit(ev)
		}
		err := p.TranslateStream(ctx, req, trackedEmit)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if produced || !IsTransient(err) || attempt >= maxRetries {
			return err
		}
		if emitErr := emit(&StreamEvent{
			Type:    EventStreamRetry,
			Attempt: attempt + 1,
			Message: Truncate(err.Error(), 120),
		}); emitErr != nil {
			return emitErr
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(streamRetryBackoff(attempt + 1)):
		}
	}
}
