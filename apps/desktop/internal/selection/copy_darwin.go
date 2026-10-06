//go:build darwin

package selection

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics
#include "selection_darwin.h"
*/
import "C"

import (
	"errors"
	"os/exec"

	"tako/internal/clipboardx"
)

// sysBackend macOS 生产后端：clipboardx 剪贴板 + CGEvent ⌘C。
type sysBackend struct{}

func (sysBackend) TakeSnapshot() (*clipboardx.Snapshot, error) { return clipboardx.TakeSnapshot() }
func (sysBackend) Restore(s *clipboardx.Snapshot) error        { return clipboardx.Restore(s) }
func (sysBackend) ReadText() (string, error)                   { return clipboardx.ReadText() }
func (sysBackend) ChangeCount() int64                          { return clipboardx.ChangeCount() }
func (sysBackend) SimulateCopy() error {
	if C.tako_simulate_copy() != 0 {
		return errors.New("CGEvent 模拟 ⌘C 失败")
	}
	return nil
}

// AccessibilityTrusted 辅助功能权限检测。prompt=true 时触发系统授权弹窗
// （bridge.CheckAccessibility 用 false 纯检测）。
func AccessibilityTrusted(prompt bool) bool {
	p := 0
	if prompt {
		p = 1
	}
	return C.tako_ax_trusted(C.int(p)) != 0
}

// OpenAccessibilitySettings 打开系统「隐私与安全性 → 辅助功能」面板。
func OpenAccessibilitySettings() error {
	return exec.Command("open",
		"x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Start()
}
