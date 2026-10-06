// clipboardx darwin 实现（Objective-C）：NSPasteboard 文本读写、变更计数、
// 全量快照/恢复。快照格式（小端未指定，进程内自产自销）：
//   u64 itemCount
//   per item: u64 typeCount
//     per type: u64 typeLen + typeBytes(UTF-8 UTI) + u64 dataLen + dataBytes
#ifndef TAKO_CLIPBOARDX_DARWIN_H
#define TAKO_CLIPBOARDX_DARWIN_H

#include <stddef.h>

typedef struct {
	void *data;
	size_t len;
} tako_pb_buffer;

long tako_pb_change_count(void);
char *tako_pb_read_text(void);            // 返回值需 free(3)；无文本返回 NULL
int tako_pb_write_text(const char *text); // 0 成功 / 非 0 失败
tako_pb_buffer tako_pb_snapshot(void);    // data 需 free(3)
int tako_pb_restore(const void *data, size_t len);

#endif
