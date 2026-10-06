package provider

import (
	"context"
	"testing"
	"time"
)

// fakeProvider 脚本化假 provider：每次调用按 run 脚本执行。
type fakeProvider struct {
	calls int
	run   func(call int, emit func(*StreamEvent) error) error
}

func (f *fakeProvider) Name() string { return "fake" }
func (f *fakeProvider) TranslateStream(_ context.Context, _ TextReq, emit func(*StreamEvent) error) error {
	f.calls++
	return f.run(f.calls, emit)
}

type collectEmit struct{ events []*StreamEvent }

func (c *collectEmit) emit(ev *StreamEvent) error {
	c.events = append(c.events, ev)
	return nil
}

func TestStreamWithRetry_RetryOnZeroOutput(t *testing.T) {
	oldBase := streamRetryBase
	streamRetryBase = time.Millisecond
	defer func() { streamRetryBase = oldBase }()

	fp := &fakeProvider{run: func(call int, emit func(*StreamEvent) error) error {
		if call < 3 {
			return &ProviderError{Code: CodeConnect, Detail: "reset"}
		}
		_ = emit(&StreamEvent{Type: EventMessageStart})
		_ = emit(&StreamEvent{Type: EventDelta, Text: "ok"})
		_ = emit(&StreamEvent{Type: EventFinish, Text: "ok"})
		return nil
	}}

	c := &collectEmit{}
	err := StreamWithRetry(context.Background(), fp,
		TextReq{Text: "x", TargetLang: "en"}, c.emit, 2)
	if err != nil {
		t.Fatalf("StreamWithRetry: %v", err)
	}
	if fp.calls != 3 {
		t.Fatalf("calls=%d, want 3", fp.calls)
	}
	var retries []int
	for _, ev := range c.events {
		if ev.Type == EventStreamRetry {
			retries = append(retries, ev.Attempt)
		}
	}
	if len(retries) != 2 || retries[0] != 1 || retries[1] != 2 {
		t.Fatalf("stream-retry attempts=%v, want [1 2]", retries)
	}
}

func TestStreamWithRetry_NoRetryAfterProduced(t *testing.T) {
	oldBase := streamRetryBase
	streamRetryBase = time.Millisecond
	defer func() { streamRetryBase = oldBase }()

	fp := &fakeProvider{run: func(call int, emit func(*StreamEvent) error) error {
		_ = emit(&StreamEvent{Type: EventDelta, Text: "partial"})
		return &ProviderError{Code: CodeConnect, Detail: "reset mid-stream"}
	}}

	c := &collectEmit{}
	err := StreamWithRetry(context.Background(), fp,
		TextReq{Text: "x", TargetLang: "en"}, c.emit, 2)
	if err == nil {
		t.Fatal("want error")
	}
	if fp.calls != 1 {
		t.Fatalf("calls=%d, want 1（已产出不重试）", fp.calls)
	}
	for _, ev := range c.events {
		if ev.Type == EventStreamRetry {
			t.Fatal("已产出不应发 stream-retry")
		}
	}
}

func TestStreamWithRetry_NoRetryOnPermanent(t *testing.T) {
	fp := &fakeProvider{run: func(call int, emit func(*StreamEvent) error) error {
		return &ProviderError{Code: CodeAuth, Detail: "bad key"}
	}}

	c := &collectEmit{}
	err := StreamWithRetry(context.Background(), fp,
		TextReq{Text: "x", TargetLang: "en"}, c.emit, 2)
	if ErrorCodeOf(err) != CodeAuth {
		t.Fatalf("code=%s, want auth", ErrorCodeOf(err))
	}
	if fp.calls != 1 {
		t.Fatalf("calls=%d, want 1", fp.calls)
	}
}

func TestStreamWithRetry_CancelDuringBackoff(t *testing.T) {
	oldBase := streamRetryBase
	streamRetryBase = time.Hour // 退避挂起，靠 ctx 解除
	defer func() { streamRetryBase = oldBase }()

	fp := &fakeProvider{run: func(call int, emit func(*StreamEvent) error) error {
		return &ProviderError{Code: CodeRateLimit, Detail: "429"}
	}}

	ctx, cancel := context.WithCancel(context.Background())
	c := &collectEmit{}
	done := make(chan error, 1)
	go func() {
		done <- StreamWithRetry(ctx, fp, TextReq{Text: "x", TargetLang: "en"}, c.emit, 2)
	}()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if ErrorCodeOf(err) != CodeCancelled {
			t.Fatalf("code=%s, want cancelled", ErrorCodeOf(err))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancel 未解除退避")
	}
}

func TestStreamWithRetry_Exhausted(t *testing.T) {
	oldBase := streamRetryBase
	streamRetryBase = time.Millisecond
	defer func() { streamRetryBase = oldBase }()

	fp := &fakeProvider{run: func(call int, emit func(*StreamEvent) error) error {
		return &ProviderError{Code: CodeTimeout, Detail: "timeout"}
	}}

	c := &collectEmit{}
	err := StreamWithRetry(context.Background(), fp,
		TextReq{Text: "x", TargetLang: "en"}, c.emit, 2)
	if ErrorCodeOf(err) != CodeTimeout {
		t.Fatalf("code=%s, want timeout", ErrorCodeOf(err))
	}
	if fp.calls != 3 {
		t.Fatalf("calls=%d, want 3（1+2 重试）", fp.calls)
	}
}
