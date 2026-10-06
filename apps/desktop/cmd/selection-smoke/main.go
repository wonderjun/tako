// selection-smoke 取词链路手动冒烟（SPEC-A A4 验收）：
//
//	go run ./cmd/selection-smoke
//
// 流程：打印辅助功能权限 → 等 5s 让用户在任意 App 选中文本 → 执行
// Capture → 打印取到的文本 + 验证剪贴板已恢复。
package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"tako/internal/clipboardx"
	"tako/internal/selection"
)

func main() {
	fmt.Printf("平台: %s\n", runtime.GOOS)
	fmt.Printf("辅助功能权限: %v\n", selection.AccessibilityTrusted(false))
	if runtime.GOOS == "darwin" && !selection.AccessibilityTrusted(false) {
		fmt.Println("未授权——执行 selection.OpenAccessibilitySettings() 打开设置面板……")
		_ = selection.OpenAccessibilitySettings()
		fmt.Println("授权后请重新运行本程序。")
		os.Exit(1)
	}

	before, _ := clipboardx.ReadText()
	fmt.Printf("取词前剪贴板文本: %q\n", before)
	fmt.Println("请在 5 秒内于任意应用选中一段文本……")
	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	start := time.Now()
	text, err := selection.Capture(ctx)
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("取词失败（耗时 %v）: %v\n", elapsed, err)
		os.Exit(1)
	}
	fmt.Printf("取词成功（耗时 %v）: %q\n", elapsed, text)

	after, _ := clipboardx.ReadText()
	fmt.Printf("取词后剪贴板文本: %q（恢复原值=%v）\n", after, after == before)
}
