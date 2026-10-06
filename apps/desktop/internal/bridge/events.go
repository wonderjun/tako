package bridge

// Go 侧事件名常量——镜像 packages/contracts/src/index.ts 的 BridgeEvents
// （契约唯一真源，禁止单边新增/改名）。
const (
	// eventTranslateEvent payload: provider.StreamEvent。
	eventTranslateEvent = "translate:event"
	// eventSelectionCaptured payload: { text: string }（M4 划词唤起接入）。
	eventSelectionCaptured = "selection:captured"
	// eventHotkeyTriggered payload: { kind: string }（M4 全局快捷键接入）。
	eventHotkeyTriggered = "hotkey:triggered"
)

// emitEvent Go→前端事件推送出口。生产实现走 wails application 单例；
// 测试替换为收集器。wails 的 Emit 返回 bool（是否有窗口消费），事件投递
// 失败不构成业务错误，故内部忽略返回值。
type emitFunc func(name string, payload any)
