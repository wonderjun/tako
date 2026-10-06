//go:build smoke

// 真实 API 冒烟测试（可选，默认不跑）：
//   go test -tags smoke ./internal/provider/ -run Smoke -v
// DeepL 需环境变量 TAKO_SMOKE_DEEPL_KEY；google/microsoft 免 key。
package provider

import (
	"context"
	"os"
	"testing"
	"time"
)

func runSmoke(t *testing.T, p Provider, name string) {
	t.Helper()
	c := &collectEmit{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := p.TranslateStream(ctx,
		TextReq{Text: "hello world", SourceLang: "en", TargetLang: "zh-Hans"}, c.emit); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	fin := c.events[len(c.events)-1]
	if fin.Type != EventFinish || fin.Text == "" {
		t.Fatalf("%s: finish=%+v", name, fin)
	}
	t.Logf("%s 译文: %s", name, fin.Text)
}

func TestSmokeDeepL(t *testing.T) {
	key := os.Getenv("TAKO_SMOKE_DEEPL_KEY")
	if key == "" {
		t.Skip("TAKO_SMOKE_DEEPL_KEY 未设置")
	}
	p, err := NewDeepL(ProviderConfig{Type: "deepl", APIKey: key})
	if err != nil {
		t.Fatal(err)
	}
	runSmoke(t, p, "deepl")
}

func TestSmokeGoogle(t *testing.T) {
	p, err := NewGoogle(ProviderConfig{Type: "google"})
	if err != nil {
		t.Fatal(err)
	}
	runSmoke(t, p, "google")
}

func TestSmokeMicrosoft(t *testing.T) {
	p, err := NewMicrosoft(ProviderConfig{Type: "microsoft"})
	if err != nil {
		t.Fatal(err)
	}
	runSmoke(t, p, "microsoft")
}
