// Package license 预留许可证模块（商业化解锁点）。
//
// 一期为 free 桩：不校验、不联网。将来收费时替换 Verifier 实现
// （激活码/订阅校验），bridge.LicenseInfo 改为消费 Verifier 结果，
// 上层 UI 与存储结构无需变动。
package license

import "time"

// Info 许可证状态快照。
type Info struct {
	Licensed  bool
	Plan      string // free | pro | trial
	ExpiresAt time.Time
}

// Verifier 许可证校验接口（一期只有 FreeVerifier 实现）。
type Verifier interface {
	// Verify 返回当前许可状态；离线场景应使用上次缓存结果。
	Verify() Info
}

// FreeVerifier 免费期桩：恒为 licensed free。
type FreeVerifier struct{}

// Verify 实现 Verifier。
func (FreeVerifier) Verify() Info {
	return Info{Licensed: true, Plan: "free"}
}
