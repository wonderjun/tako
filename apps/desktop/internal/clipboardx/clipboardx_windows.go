//go:build windows

// clipboardx windows 实现：user32 剪贴板 API + kernel32 Global* 内存 API。
// 快照 = EnumClipboardFormats 全格式逐字节复制（GMEM 句柄可深拷贝；
// CF_BITMAP/HPALETTE 等 GDI 对象句柄无法字节复制，拿不到数据的格式跳过）。
// 快照格式（进程内自产自销）：
//
//	u64 formatCount
//	per format: u32 formatID + u64 dataLen + dataBytes
package clipboardx

import (
	"bytes"
	"encoding/binary"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	cfUnicodeText = 13 // CF_UNICODETEXT
	gmemMoveable  = 0x0002
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	procOpenClipboard              = user32.NewProc("OpenClipboard")
	procCloseClipboard             = user32.NewProc("CloseClipboard")
	procEmptyClipboard             = user32.NewProc("EmptyClipboard")
	procEnumClipboardFormats       = user32.NewProc("EnumClipboardFormats")
	procGetClipboardData           = user32.NewProc("GetClipboardData")
	procSetClipboardData           = user32.NewProc("SetClipboardData")
	procGetClipboardSequenceNumber = user32.NewProc("GetClipboardSequenceNumber")

	procGlobalAlloc   = kernel32.NewProc("GlobalAlloc")
	procGlobalLock    = kernel32.NewProc("GlobalLock")
	procGlobalUnlock  = kernel32.NewProc("GlobalUnlock")
	procGlobalSize    = kernel32.NewProc("GlobalSize")
	procGlobalFree    = kernel32.NewProc("GlobalFree")
	procRtlMoveMemory = kernel32.NewProc("RtlMoveMemory")
)

func withClipboard(fn func() error) error {
	// OpenClipboard 需短暂重试：别的进程可能正持有（竞态窗口内常见）。
	var lastErr error
	for i := 0; i < 10; i++ {
		r, _, err := procOpenClipboard.Call(0)
		if r != 0 {
			defer procCloseClipboard.Call()
			return fn()
		}
		lastErr = err
		windows.SleepEx(5, false)
	}
	return lastErr
}

func readText() (string, error) {
	var out string
	err := withClipboard(func() error {
		r, _, err := procGetClipboardData.Call(cfUnicodeText)
		if r == 0 {
			return nil // 无文本内容不算错误
		}
		p, _, err := procGlobalLock.Call(r)
		if p == 0 {
			return err
		}
		defer procGlobalUnlock.Call(r)
		out = utf16PtrToString((*uint16)(unsafe.Pointer(p)))
		return nil
	})
	return out, err
}

// utf16PtrToString 读取 NUL 结尾的 UTF-16 缓冲区。
func utf16PtrToString(p *uint16) string {
	var n int
	for *(*uint16)(unsafe.Pointer(uintptr(unsafe.Pointer(p)) + uintptr(n)*2)) != 0 {
		n++
	}
	u16 := unsafe.Slice(p, n)
	return string(utf16.Decode(u16))
}

func writeText(text string) error {
	u16, err := windows.UTF16FromString(text)
	if err != nil {
		return err
	}
	size := uintptr(len(u16) * 2)
	return withClipboard(func() error {
		if r, _, err := procEmptyClipboard.Call(); r == 0 {
			return err
		}
		h, _, err := procGlobalAlloc.Call(gmemMoveable, size)
		if h == 0 {
			return err
		}
		p, _, err := procGlobalLock.Call(h)
		if p == 0 {
			procGlobalFree.Call(h)
			return err
		}
		procRtlMoveMemory.Call(p, uintptr(unsafe.Pointer(&u16[0])), size)
		procGlobalUnlock.Call(h)
		if r, _, err := procSetClipboardData.Call(cfUnicodeText, h); r == 0 {
			procGlobalFree.Call(h)
			return err
		}
		return nil
	})
}

func takeSnapshot() (*Snapshot, error) {
	var buf bytes.Buffer
	var countPh = make([]byte, 8)
	buf.Write(countPh) // 占位，末尾回填 formatCount
	var formatCount uint64

	err := withClipboard(func() error {
		format := uintptr(0)
		for {
			r, _, _ := procEnumClipboardFormats.Call(format)
			if r == 0 {
				break // 枚举完毕
			}
			format = r
			h, _, _ := procGetClipboardData.Call(format)
			if h == 0 {
				continue // 该格式当前不可取（如 promised 数据），跳过
			}
			size, _, _ := procGlobalSize.Call(h)
			if size == 0 {
				continue // 非 GMEM 句柄（GDI 对象等），跳过
			}
			p, _, _ := procGlobalLock.Call(h)
			if p == 0 {
				continue
			}
			data := make([]byte, int(size))
			procRtlMoveMemory.Call(uintptr(unsafe.Pointer(&data[0])), p, size)
			procGlobalUnlock.Call(h)

			var hdr [12]byte
			binary.LittleEndian.PutUint32(hdr[0:4], uint32(format))
			binary.LittleEndian.PutUint64(hdr[4:12], uint64(len(data)))
			buf.Write(hdr[:])
			buf.Write(data)
			formatCount++
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	raw := buf.Bytes()
	binary.LittleEndian.PutUint64(raw[0:8], formatCount)
	return &Snapshot{raw: raw}, nil
}

func restore(s *Snapshot) error {
	r := bytes.NewReader(s.raw)
	var formatCount uint64
	if err := binary.Read(r, binary.LittleEndian, &formatCount); err != nil {
		return err
	}
	type entry struct {
		format uint32
		data   []byte
	}
	entries := make([]entry, 0, formatCount)
	for i := uint64(0); i < formatCount; i++ {
		var format uint32
		var size uint64
		if err := binary.Read(r, binary.LittleEndian, &format); err != nil {
			return err
		}
		if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
			return err
		}
		data := make([]byte, int(size))
		if _, err := r.Read(data); err != nil {
			return err
		}
		entries = append(entries, entry{format: format, data: data})
	}

	return withClipboard(func() error {
		if ret, _, err := procEmptyClipboard.Call(); ret == 0 {
			return err
		}
		for _, e := range entries {
			size := uintptr(len(e.data))
			h, _, err := procGlobalAlloc.Call(gmemMoveable, size)
			if h == 0 {
				return err
			}
			p, _, err := procGlobalLock.Call(h)
			if p == 0 {
				procGlobalFree.Call(h)
				return err
			}
			procRtlMoveMemory.Call(p, uintptr(unsafe.Pointer(&e.data[0])), size)
			procGlobalUnlock.Call(h)
			if ret, _, err := procSetClipboardData.Call(uintptr(e.format), h); ret == 0 {
				procGlobalFree.Call(h)
				return err
			}
		}
		return nil
	})
}

func changeCount() int64 {
	r, _, _ := procGetClipboardSequenceNumber.Call()
	return int64(r)
}
