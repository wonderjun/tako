// Package selection 划词取词：模拟复制按键 → 轮询剪贴板新值 → 恢复剪贴板。
//
// 平台分层：copy_darwin.go（CGEvent 模拟 ⌘C + AXIsProcessTrustedWithOptions
// 权限检测）、copy_windows.go（SendInput 模拟 Ctrl+C）、copy_stub.go。
// 状态机本体（capture）平台无关，Backend 注入便于单测。
package selection

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tako/internal/clipboardx"
)

// ErrNoSelection 未取到文本：无选中内容、目标应用不响应模拟复制、或超时。
var ErrNoSelection = errors.New("selection: 未取到选中文本")

// Backend 取词依赖的平台能力（剪贴板读写快照 + 模拟复制按键）。
// 生产实现见 copy_*.go 的 sysBackend；测试注入假实现。
type Backend interface {
	TakeSnapshot() (*clipboardx.Snapshot, error)
	Restore(*clipboardx.Snapshot) error
	ReadText() (string, error)
	ChangeCount() int64
	SimulateCopy() error
}

// 取词时序预算（SPEC-A：等新值 300ms 超时、轮询探测间隔 15ms）。
var (
	pollInterval = 15 * time.Millisecond
	pollTimeout  = 300 * time.Millisecond
)

// Capture 执行一次完整取词。任何路径（含出错）都会恢复用户剪贴板。
// ctx 取消立即返回（恢复仍执行）。
func Capture(ctx context.Context) (string, error) {
	return capture(ctx, sysBackend{})
}

// capture 状态机（Backend 注入便于单测）：
//
//	快照 → 记录 changeCount → 模拟复制 → 轮询 changeCount 变化
//	→ 读文本 →（defer）恢复快照 → 返回文本
func capture(ctx context.Context, b Backend) (string, error) {
	snap, err := b.TakeSnapshot()
	if err != nil {
		return "", fmt.Errorf("selection: 剪贴板快照失败: %w", err)
	}
	defer func() { _ = b.Restore(snap) }()

	prevCount := b.ChangeCount()

	if err := b.SimulateCopy(); err != nil {
		return "", fmt.Errorf("selection: 模拟复制失败: %w", err)
	}

	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		if b.ChangeCount() != prevCount {
			break
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(pollInterval):
		}
	}
	if b.ChangeCount() == prevCount {
		return "", ErrNoSelection
	}

	text, err := b.ReadText()
	if err != nil {
		return "", fmt.Errorf("selection: 读取剪贴板失败: %w", err)
	}
	if text == "" {
		return "", ErrNoSelection
	}
	return text, nil
}
