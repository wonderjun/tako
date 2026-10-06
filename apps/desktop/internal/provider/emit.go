package provider

// emitOnce 传统非流式服务的统一事件序列：message-start → delta{全量} →
// finish{全量}。非流式服务拿到完整响应后一次性回放（时序契约见
// packages/contracts README「流式事件时序」）。
func emitOnce(emit func(*StreamEvent) error, text string) error {
	if err := emit(&StreamEvent{Type: EventMessageStart}); err != nil {
		return err
	}
	if err := emit(&StreamEvent{Type: EventDelta, Text: text}); err != nil {
		return err
	}
	return emit(&StreamEvent{Type: EventFinish, Text: text})
}
