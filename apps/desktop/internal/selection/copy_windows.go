//go:build windows

// selection windows 实现：SendInput 模拟 Ctrl+C（4 个键事件：
// Ctrl down → C down → C up → Ctrl up）。Windows 无需辅助功能权限。
package selection

import (
	"errors"
	"unsafe"

	"tako/internal/clipboardx"

	"golang.org/x/sys/windows"
)

const (
	inputKeyboard  = 1
	keyeventfKeyup = 0x0002
	vkControl      = 0x11
	vkC            = 0x43
)

// INPUT/KEYBDINPUT 结构体（Win32）：INPUT 在 64 位下 40 字节
// （type 4 + padding 4 + union 32），union 按最大的 MOUSEINPUT 补齐。
type keybdInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

type input struct {
	typ  uint32
	_    uint32 // union 对齐 padding
	ki   keybdInput
	_    [8]byte // union 剩余（32 - sizeof(keybdInput)=24）
}

var procSendInput = windows.NewLazySystemDLL("user32.dll").NewProc("SendInput")

func simulateCopy() error {
	inputs := [4]input{
		{typ: inputKeyboard, ki: keybdInput{wVk: vkControl}},
		{typ: inputKeyboard, ki: keybdInput{wVk: vkC}},
		{typ: inputKeyboard, ki: keybdInput{wVk: vkC, dwFlags: keyeventfKeyup}},
		{typ: inputKeyboard, ki: keybdInput{wVk: vkControl, dwFlags: keyeventfKeyup}},
	}
	r, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if r != uintptr(len(inputs)) {
		if err != nil && err != windows.ERROR_SUCCESS {
			return err
		}
		return errors.New("SendInput 未发送全部键事件")
	}
	return nil
}

// sysBackend windows 生产后端：clipboardx 剪贴板 + SendInput Ctrl+C。
type sysBackend struct{}

func (sysBackend) TakeSnapshot() (*clipboardx.Snapshot, error) { return clipboardx.TakeSnapshot() }
func (sysBackend) Restore(s *clipboardx.Snapshot) error        { return clipboardx.Restore(s) }
func (sysBackend) ReadText() (string, error)                   { return clipboardx.ReadText() }
func (sysBackend) ChangeCount() int64                          { return clipboardx.ChangeCount() }
func (sysBackend) SimulateCopy() error                         { return simulateCopy() }

// AccessibilityTrusted Windows 无需辅助功能权限，恒 true。
func AccessibilityTrusted(bool) bool { return true }

// OpenAccessibilitySettings Windows 无障碍面板无对应概念，no-op。
func OpenAccessibilitySettings() error { return nil }
