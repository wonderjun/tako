// selection darwin 实现（cgo 编译）。
// CGEventPost 到 HID tap 需要辅助功能权限（AXIsProcessTrustedWithOptions 检测）。
#import "selection_darwin.h"

#import <ApplicationServices/ApplicationServices.h>
#import <CoreGraphics/CoreGraphics.h>

int tako_ax_trusted(int prompt) {
	const void *keys[] = {kAXTrustedCheckOptionPrompt};
	const void *vals[] = {prompt ? kCFBooleanTrue : kCFBooleanFalse};
	CFDictionaryRef opts = CFDictionaryCreate(
		NULL, keys, vals, 1,
		&kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
	Boolean trusted = AXIsProcessTrustedWithOptions(opts);
	CFRelease(opts);
	return trusted ? 1 : 0;
}

// kVK_ANSI_C = 8（macOS 虚拟键码，C 键）。
int tako_simulate_copy(void) {
	CGEventSourceRef src = CGEventSourceCreate(kCGEventSourceStateHIDSystemState);
	if (!src) return -1;
	CGEventRef down = CGEventCreateKeyboardEvent(src, 8, true);
	CGEventRef up = CGEventCreateKeyboardEvent(src, 8, false);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up) CFRelease(up);
		CFRelease(src);
		return -2;
	}
	CGEventSetFlags(down, kCGEventFlagMaskCommand);
	CGEventSetFlags(up, kCGEventFlagMaskCommand);
	CGEventPost(kCGHIDEventTap, down);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	CFRelease(src);
	return 0;
}
