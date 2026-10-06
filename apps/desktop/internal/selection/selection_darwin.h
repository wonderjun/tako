// selection darwin 系统调用（Objective-C / CoreGraphics C API）：
// AXIsProcessTrustedWithOptions 权限检测、CGEvent 模拟 ⌘C。
#ifndef TAKO_SELECTION_DARWIN_H
#define TAKO_SELECTION_DARWIN_H

int tako_ax_trusted(int prompt);   // 1=已授权 / 0=未授权；prompt=1 时触发系统弹窗
int tako_simulate_copy(void);      // 0 成功 / 非 0 失败

#endif
