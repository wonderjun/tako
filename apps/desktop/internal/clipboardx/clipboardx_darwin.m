// clipboardx darwin 实现（Objective-C，cgo 编译）。
// 注意：快照里 dataForType 可能为 nil（promised/lazy 类型如文件拖放）——
// 该 type 恢复时无数据可写，直接跳过，不阻断其余类型。
#import "clipboardx_darwin.h"

#import <AppKit/AppKit.h>
#import <Foundation/Foundation.h>
#include <stdlib.h>
#include <string.h>

static void buf_write_u64(NSMutableData *buf, uint64_t v) {
	[buf appendBytes:&v length:sizeof(v)];
}

static void buf_write(NSMutableData *buf, const void *p, uint64_t n) {
	[buf appendBytes:p length:(NSUInteger)n];
}

// 游标式解析（写入方即本进程，格式自洽，越界防御即可）。
typedef struct {
	const uint8_t *p;
	const uint8_t *end;
} cursor;

static int cur_read_u64(cursor *c, uint64_t *out) {
	if ((size_t)(c->end - c->p) < sizeof(uint64_t)) return -1;
	memcpy(out, c->p, sizeof(uint64_t));
	c->p += sizeof(uint64_t);
	return 0;
}

static int cur_read_bytes(cursor *c, const uint8_t **p, uint64_t n) {
	if ((uint64_t)(c->end - c->p) < n) return -1;
	*p = c->p;
	c->p += n;
	return 0;
}

long tako_pb_change_count(void) {
	return (long)[[NSPasteboard generalPasteboard] changeCount];
}

char *tako_pb_read_text(void) {
	NSString *s = [[NSPasteboard generalPasteboard] stringForType:NSPasteboardTypeString];
	if (!s) return NULL;
	return strdup([s UTF8String]);
}

int tako_pb_write_text(const char *text) {
	NSPasteboard *pb = [NSPasteboard generalPasteboard];
	[pb clearContents];
	NSString *s = [NSString stringWithUTF8String:text];
	if (!s) return -1;
	return [pb setString:s forType:NSPasteboardTypeString] ? 0 : -1;
}

tako_pb_buffer tako_pb_snapshot(void) {
	tako_pb_buffer out = {NULL, 0};
	NSPasteboard *pb = [NSPasteboard generalPasteboard];
	NSArray<NSPasteboardItem *> *items = [pb pasteboardItems];
	NSMutableData *buf = [NSMutableData data];

	buf_write_u64(buf, (uint64_t)items.count);
	for (NSPasteboardItem *item in items) {
		NSArray<NSString *> *types = [item types];
		buf_write_u64(buf, (uint64_t)types.count);
		for (NSString *t in types) {
			NSData *td = [t dataUsingEncoding:NSUTF8StringEncoding];
			buf_write_u64(buf, (uint64_t)td.length);
			buf_write(buf, td.bytes, td.length);
			NSData *d = [item dataForType:t];
			uint64_t dl = d ? (uint64_t)d.length : 0;
			buf_write_u64(buf, dl);
			if (dl > 0) buf_write(buf, d.bytes, dl);
		}
	}

	out.len = buf.length;
	out.data = malloc(out.len > 0 ? out.len : 1);
	if (out.data && out.len > 0) memcpy(out.data, buf.bytes, out.len);
	return out;
}

int tako_pb_restore(const void *data, size_t len) {
	if (!data || len == 0) return -1;
	cursor c = {(const uint8_t *)data, (const uint8_t *)data + len};

	uint64_t itemCount = 0;
	if (cur_read_u64(&c, &itemCount) != 0) return -1;

	NSMutableArray<NSPasteboardItem *> *arr = [NSMutableArray arrayWithCapacity:(NSUInteger)itemCount];
	for (uint64_t i = 0; i < itemCount; i++) {
		uint64_t typeCount = 0;
		if (cur_read_u64(&c, &typeCount) != 0) return -1;
		NSPasteboardItem *item = [[NSPasteboardItem alloc] init];
		for (uint64_t j = 0; j < typeCount; j++) {
			uint64_t tl = 0, dl = 0;
			const uint8_t *tp = NULL, *dp = NULL;
			if (cur_read_u64(&c, &tl) != 0 || cur_read_bytes(&c, &tp, tl) != 0 ||
				cur_read_u64(&c, &dl) != 0) {
				return -1;
			}
			// dataLen=0（promised/lazy 类型）恢复时跳过该 type。
			if (dl > 0 && cur_read_bytes(&c, &dp, dl) != 0) return -1;
			NSString *type = [[NSString alloc] initWithBytes:tp length:(NSUInteger)tl
												encoding:NSUTF8StringEncoding];
			if (!type || dl == 0) continue;
			NSData *d = [NSData dataWithBytes:dp length:(NSUInteger)dl];
			[item setData:d forType:type];
		}
		[arr addObject:item];
	}

	NSPasteboard *pb = [NSPasteboard generalPasteboard];
	[pb clearContents];
	return [pb writeObjects:arr] ? 0 : -1;
}
