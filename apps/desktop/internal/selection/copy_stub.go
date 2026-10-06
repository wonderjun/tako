//go:build !darwin && !windows

package selection

import (
	"errors"

	"tako/internal/clipboardx"
)

// sysBackend 非支持平台 stub：Capture 会因 SimulateCopy 报错。
type sysBackend struct{}

func (sysBackend) TakeSnapshot() (*clipboardx.Snapshot, error) {
	return nil, clipboardx.ErrUnsupported
}
func (sysBackend) Restore(*clipboardx.Snapshot) error  { return clipboardx.ErrUnsupported }
func (sysBackend) ReadText() (string, error)           { return "", clipboardx.ErrUnsupported }
func (sysBackend) ChangeCount() int64                  { return 0 }
func (sysBackend) SimulateCopy() error                 { return errors.New("selection: unsupported platform") }

// AccessibilityTrusted 非支持平台恒 true（不阻塞流程）。
func AccessibilityTrusted(bool) bool { return true }

// OpenAccessibilitySettings 非支持平台 no-op。
func OpenAccessibilitySettings() error { return nil }
