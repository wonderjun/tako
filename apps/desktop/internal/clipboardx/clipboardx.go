// Package clipboardx 剪贴板读写与全量快照/恢复（取词流程的基础设施）。
//
// 取词需要模拟 ⌘C/Ctrl+C 把选中文本写进剪贴板——这会覆盖用户原有内容。
// 快照/恢复让整条链路对用户无感：Snapshot → 模拟复制 → 读新值 → Restore。
//
// Snapshot 的序列化形态各平台自定义为 []byte（内容不透明）：
//   - darwin：NSPasteboard 全部 pasteboardItem × type → data
//   - windows：EnumClipboardFormats 全格式 × GlobalLock 字节
//
// 平台实现在 clipboardx_{darwin,windows}.go，其余平台见 stub（报
// ErrUnsupported）。
package clipboardx

import "errors"

// ErrUnsupported 当前平台不支持剪贴板快照（非 darwin/windows）。
var ErrUnsupported = errors.New("clipboardx: unsupported platform")

// Snapshot 剪贴板全量快照（不透明的序列化字节，平台实现自定义格式）。
// 恢复语义：Restore(s) 将剪贴板完全还原为快照时刻的内容。
type Snapshot struct {
	raw []byte
}

// ReadText 读取剪贴板纯文本（无文本返回空串+nil）。
func ReadText() (string, error) { return readText() }

// WriteText 覆盖写入纯文本。
func WriteText(text string) error { return writeText(text) }

// TakeSnapshot 快照当前剪贴板全部内容。
func TakeSnapshot() (*Snapshot, error) { return takeSnapshot() }

// Restore 将剪贴板恢复为快照内容；nil 快照为 no-op。
func Restore(s *Snapshot) error {
	if s == nil {
		return nil
	}
	return restore(s)
}

// ChangeCount 剪贴板变更计数（darwin: NSPasteboard.changeCount；
// windows: GetClipboardSequenceNumber）。用于探测"模拟复制是否产出了新值"。
func ChangeCount() int64 { return changeCount() }
