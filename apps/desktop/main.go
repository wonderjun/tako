// Tako 桌面应用入口：窗口注册、托盘菜单与服务装配。
// 窗口三件套：main（设置/历史，关窗隐藏到托盘）、popup（划词浮窗，失焦隐藏）、
// input（输入翻译，托盘/快捷键唤起）。前端路由为 hash 模式，URL 以 /#/ 区分页面。
package main

import (
	"embed"
	"log"

	"tako/internal/bridge"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"github.com/wailsapp/wails/v3/pkg/icons"
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
			// 托盘常驻：关窗不退出，由托盘/点击 Dock 图标唤起
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	// 主窗口：设置 / 服务管理 / 历史。紧凑型工具窗（对标 Bob）：
	// 限最大宽度防超宽拉伸，内容区各页自带适度 max-w。
	mainWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:      "main",
		Title:     "Tako",
		Width:     960,
		Height:    640,
		MinWidth:  720,
		MinHeight: 480,
		MaxWidth:  1100,
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

	// 输入窗：⌥A / 托盘菜单唤起的快捷输入翻译。
	inputWindow := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:            "input",
		Width:           560,
		Height:          420,
		URL:             "/#/input",
		Hidden:          true,
		AlwaysOnTop:     true,
		HideOnFocusLost: true,
	})

	// 关闭主窗口 = 隐藏到托盘（退出走托盘菜单）。
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		mainWindow.Hide()
		e.Cancel()
	})

	// 点击 Dock 图标重新显示主窗口（macOS 惯例）。
	app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(*application.ApplicationEvent) {
		mainWindow.Show()
	})

	// ── 托盘：常驻菜单（对标 Bob 的菜单栏入口） ──
	tray := app.SystemTray.New()
	tray.SetTooltip("Tako · 章鱼")
	tray.SetTemplateIcon(icons.SystrayMacTemplate)

	trayMenu := app.NewMenu()
	trayMenu.Add("显示 / 隐藏主窗口").OnClick(func(*application.Context) {
		if mainWindow.IsVisible() {
			mainWindow.Hide()
		} else {
			mainWindow.Show()
		}
	})
	trayMenu.AddSeparator()
	trayMenu.Add("输入翻译").OnClick(func(*application.Context) {
		// 唤起输入窗并通知前端聚焦输入框（B5 已接 hotkey:triggered）
		inputWindow.Show()
		app.Event.Emit("hotkey:triggered", map[string]string{"kind": "input"})
	})
	trayMenu.AddSeparator()
	trayMenu.Add("偏好设置…").OnClick(func(*application.Context) {
		mainWindow.Show()
		mainWindow.SetURL("/#/settings")
	})
	trayMenu.Add("退出 Tako").OnClick(func(*application.Context) {
		app.Quit()
	})
	tray.SetMenu(trayMenu)
	// 托盘图标左键点击 = 显示主窗口（菜单从右键/长按出）
	tray.OnClick(func() { mainWindow.Show() })

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
