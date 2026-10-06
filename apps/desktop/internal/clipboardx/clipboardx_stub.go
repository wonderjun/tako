//go:build !darwin && !windows

package clipboardx

// 非支持平台 stub：全部报 ErrUnsupported（保证 linux 等可编译）。

func readText() (string, error)        { return "", ErrUnsupported }
func writeText(string) error           { return ErrUnsupported }
func takeSnapshot() (*Snapshot, error) { return nil, ErrUnsupported }
func restore(*Snapshot) error          { return ErrUnsupported }
func changeCount() int64               { return 0 }
