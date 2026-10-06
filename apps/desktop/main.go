// Tako 桌面应用入口：窗口注册与服务装配。
// 窗口三件套：main（设置/历史）、popup（划词浮窗，失焦隐藏）、input（输入翻译）。
// 前端路由为 hash 模式，窗口 URL 以 /#/ 前缀区分页面。
package main

import (
	"embed"
	"log"

	"tako/internal/bridge"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Tako",
		Description: "多触手聚合翻译：划词即译，即用即走",
		Services: []application.Service{
			application.NewService(bridge.NewService()),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// 托盘常驻：关窗不退出，由托盘/快捷键唤起
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 主窗口：设置 / 服务管理 / 历史
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      "main",
		Title:     "Tako",
		Width:     960,
		Height:    640,
		MinWidth:  720,
		MinHeight: 480,
		URL:       "/#/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	// 浮窗：划词/截图翻译结果展示。frameless + 透明背景 + 失焦自动隐藏，
	// 由 selection 取词完成后经 selection:captured 事件唤起（M4 接入）。
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "popup",
		Width:           420,
		Height:          300,
		URL:             "/#/popup",
		Frameless:       true,
		Hidden:          true,
		AlwaysOnTop:     true,
		HideOnFocusLost: true,
		BackgroundType:  application.BackgroundTypeTransparent,
	})

	// 输入窗：⌥A 唤起的快捷输入翻译。
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "input",
		Width:           560,
		Height:          420,
		URL:             "/#/input",
		Hidden:          true,
		AlwaysOnTop:     true,
		HideOnFocusLost: true,
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
