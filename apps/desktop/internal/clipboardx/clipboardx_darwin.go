//go:build darwin

package clipboardx

/*
#cgo LDFLAGS: -framework AppKit -framework Foundation
#include <stdlib.h>
#include "clipboardx_darwin.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

func readText() (string, error) {
	p := C.tako_pb_read_text()
	if p == nil {
		return "", nil
	}
	defer C.free(unsafe.Pointer(p))
	return C.GoString(p), nil
}

func writeText(text string) error {
	cstr := C.CString(text)
	defer C.free(unsafe.Pointer(cstr))
	if C.tako_pb_write_text(cstr) != 0 {
		return errors.New("clipboardx: writeText 失败")
	}
	return nil
}

func takeSnapshot() (*Snapshot, error) {
	buf := C.tako_pb_snapshot()
	if buf.data == nil {
		return nil, errors.New("clipboardx: 快照失败（pasteboardItems 不可用）")
	}
	defer C.free(buf.data)
	return &Snapshot{raw: C.GoBytes(buf.data, C.int(buf.len))}, nil
}

func restore(s *Snapshot) error {
	if len(s.raw) == 0 {
		return nil
	}
	if C.tako_pb_restore(unsafe.Pointer(&s.raw[0]), C.size_t(len(s.raw))) != 0 {
		return errors.New("clipboardx: 恢复失败")
	}
	return nil
}

func changeCount() int64 {
	return int64(C.tako_pb_change_count())
}
