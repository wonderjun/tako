package selection

import (
	"context"
	"errors"
	"testing"
	"time"

	"tako/internal/clipboardx"
)

// fakeBackend 脚本化假后端：记录调用、可控 changeCount/文本/错误。
type fakeBackend struct {
	count       int64
	text        string
	snapCalls   int
	restoreCalls int
	copyCalls   int
	copyErr     error
	readErr     error
	snapErr     error
	// copyHook 模拟复制时的副作用（如 bump count 表示剪贴板被改写）。
	copyHook func(*fakeBackend)
}

func (f *fakeBackend) TakeSnapshot() (*clipboardx.Snapshot, error) {
	f.snapCalls++
	if f.snapErr != nil {
		return nil, f.snapErr
	}
	return &clipboardx.Snapshot{}, nil
}
func (f *fakeBackend) Restore(*clipboardx.Snapshot) error {
	f.restoreCalls++
	return nil
}
func (f *fakeBackend) ReadText() (string, error) { return f.text, f.readErr }
func (f *fakeBackend) ChangeCount() int64        { return f.count }
func (f *fakeBackend) SimulateCopy() error {
	f.copyCalls++
	if f.copyErr != nil {
		return f.copyErr
	}
	if f.copyHook != nil {
		f.copyHook(f)
	}
	return nil
}

func TestCapture_Success(t *testing.T) {
	fb := &fakeBackend{
		text: "选中文本",
		copyHook: func(f *fakeBackend) { f.count++ },
	}
	got, err := capture(context.Background(), fb)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if got != "选中文本" {
		t.Fatalf("text=%q", got)
	}
	if fb.snapCalls != 1 || fb.copyCalls != 1 || fb.restoreCalls != 1 {
		t.Fatalf("calls snap=%d copy=%d restore=%d", fb.snapCalls, fb.copyCalls, fb.restoreCalls)
	}
}

func TestCapture_TimeoutNoChange(t *testing.T) {
	oldInterval, oldTimeout := pollInterval, pollTimeout
	pollInterval, pollTimeout = time.Millisecond, 30*time.Millisecond
	defer func() { pollInterval, pollTimeout = oldInterval, oldTimeout }()

	fb := &fakeBackend{} // count 不变 → 模拟复制未产出
	_, err := capture(context.Background(), fb)
	if !errors.Is(err, ErrNoSelection) {
		t.Fatalf("err=%v, want ErrNoSelection", err)
	}
	if fb.restoreCalls != 1 {
		t.Fatal("超时也须恢复剪贴板")
	}
}

func TestCapture_EmptyText(t *testing.T) {
	fb := &fakeBackend{
		text:     "",
		copyHook: func(f *fakeBackend) { f.count++ },
	}
	_, err := capture(context.Background(), fb)
	if !errors.Is(err, ErrNoSelection) {
		t.Fatalf("err=%v, want ErrNoSelection（空文本）", err)
	}
}

func TestCapture_SnapshotFailNoCopy(t *testing.T) {
	fb := &fakeBackend{snapErr: errors.New("snap fail")}
	_, err := capture(context.Background(), fb)
	if err == nil {
		t.Fatal("want error")
	}
	if fb.copyCalls != 0 {
		t.Fatal("快照失败不应触发模拟复制")
	}
}

func TestCapture_CopyFailStillRestores(t *testing.T) {
	fb := &fakeBackend{copyErr: errors.New("copy fail")}
	_, err := capture(context.Background(), fb)
	if err == nil {
		t.Fatal("want error")
	}
	if fb.restoreCalls != 1 {
		t.Fatal("模拟复制失败也须恢复剪贴板")
	}
}

func TestCapture_CtxCancelDuringPoll(t *testing.T) {
	oldInterval, oldTimeout := pollInterval, pollTimeout
	pollInterval, pollTimeout = time.Millisecond, 10*time.Second
	defer func() { pollInterval, pollTimeout = oldInterval, oldTimeout }()

	fb := &fakeBackend{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := capture(ctx, fb); done <- err }()
	time.Sleep(20 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err=%v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ctx 取消未中断轮询")
	}
	if fb.restoreCalls != 1 {
		t.Fatal("取消也须恢复剪贴板")
	}
}
