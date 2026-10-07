// 共享 HTTP 基础设施：流式 client（无总时长上限 + 首包超时 + 空闲看门狗）
// 与普通 client（有总时长上限，供传统翻译 API）。本模块保持零日志依赖。
package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// connect/read 分离语义：连接阶段快速失败（dial 10s），首包受
// ResponseHeaderTimeout 约束，流式正文不设总时长上限（长流由调用方
// ctx 与空闲看门狗兜底，不限制任务总时长）。
const (
	streamDialTimeout = 10 * time.Second
	streamKeepAlive   = 30 * time.Second

	// 首包（响应头）与空闲看门狗上限：取同值 300s，覆盖慢接入点/长 prefill
	// 与「上游整段缓冲长时间不吐字节」的场景，只防僵死不限制总时长。
	streamResponseHeaderTimeout = 300 * time.Second
	streamIdleWatchdog          = 300 * time.Second
)

// NewStreamHTTPClient 构建流式调用专用 client（各流式 provider 共用形态）。
// Client.Timeout 恒为 0：正文阶段不设总时长上限。
func NewStreamHTTPClient() *http.Client {
	dialer := &net.Dialer{Timeout: streamDialTimeout, KeepAlive: streamKeepAlive}
	return &http.Client{
		Timeout: 0,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			ResponseHeaderTimeout: streamResponseHeaderTimeout,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			ForceAttemptHTTP2:     true,
		},
	}
}

// NewHTTPClient 普通请求 client（传统翻译 API：deepl/google/microsoft），
// 有总时长上限——非流式请求整个生命周期都应快速完成或失败。
func NewHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &http.Client{Timeout: timeout}
}

// ---------------------------------------------------------------------------
// 空闲看门狗 ReadCloser
// ---------------------------------------------------------------------------

// errStreamIdleTimeout 看门狗触发的哨兵错误：NormalizeStreamError 据此归一
// 为 timeout（否则读到的是 "http: read on closed response body"，无法分辨
// 真实原因）。
var errStreamIdleTimeout = errors.New("stream idle timeout")

// idleWatchdogReadCloser 包装流式响应体：每次读到数据即重置计时器，连续
// idle 时长无数据则主动关闭底层 body 使 SSE 读取报错退出。
type idleWatchdogReadCloser struct {
	rc   io.ReadCloser
	idle time.Duration

	mu     sync.Mutex
	timer  *time.Timer
	closed bool
	fired  bool
}

// NewIdleWatchdogReadCloser 包装流式响应体挂空闲看门狗。
func NewIdleWatchdogReadCloser(rc io.ReadCloser, idle time.Duration) io.ReadCloser {
	if idle <= 0 {
		idle = streamIdleWatchdog
	}
	w := &idleWatchdogReadCloser{rc: rc, idle: idle}
	w.timer = time.AfterFunc(idle, func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		if !w.closed {
			w.fired = true
			_ = w.rc.Close()
		}
	})
	return w
}

func (w *idleWatchdogReadCloser) Read(p []byte) (int, error) {
	n, err := w.rc.Read(p)
	if n > 0 {
		w.mu.Lock()
		if !w.closed {
			w.timer.Reset(w.idle)
		}
		w.mu.Unlock()
	}
	if err != nil {
		w.mu.Lock()
		fired := w.fired
		w.mu.Unlock()
		if fired {
			// Close 解除阻塞读后的错误统一归因为看门狗触发，避免泄漏
			// "http: read on closed response body" 这类实现细节
			return n, errStreamIdleTimeout
		}
	}
	return n, err
}

func (w *idleWatchdogReadCloser) Close() error {
	w.mu.Lock()
	w.closed = true
	w.timer.Stop()
	w.mu.Unlock()
	return w.rc.Close()
}

// ---------------------------------------------------------------------------
// 流式错误归一化
// ---------------------------------------------------------------------------

// NormalizeStreamError 将请求/读流阶段的底层错误归一化为 *ProviderError。
// 调用方主动取消/自身 deadline 原样返回（由上层据此映射 cancelled）。
func NormalizeStreamError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	var pe *ProviderError
	if errors.As(err, &pe) {
		return err
	}
	if errors.Is(err, errStreamIdleTimeout) {
		return &ProviderError{
			Code:   CodeTimeout,
			Detail: fmt.Sprintf("流式空闲超时（%d 秒无数据）", int(streamIdleWatchdog.Seconds())),
			Cause:  err,
		}
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() || errors.Is(err, os.ErrDeadlineExceeded) {
		detail := "流式请求超时"
		if strings.Contains(err.Error(), "timeout awaiting response headers") {
			detail = fmt.Sprintf("首包响应超时（%d 秒未开始返回）", int(streamResponseHeaderTimeout.Seconds()))
		}
		return &ProviderError{Code: CodeTimeout, Detail: detail, Cause: err}
	}
	// 传输层失败归 connect：dial 建连失败、读流中途 reset/broken pipe、
	// 帧中途截断的 unexpected EOF。
	if IsConnectError(err) || IsStreamBreakError(err) {
		return &ProviderError{
			Code:   CodeConnect,
			Detail: fmt.Sprintf("服务连接中断: %s", err),
			Cause:  err,
		}
	}
	return &ProviderError{Code: CodeUnknown, Detail: err.Error(), Cause: err}
}
