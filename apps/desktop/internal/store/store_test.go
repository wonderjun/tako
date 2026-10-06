package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func tempStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "tako.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

// schemaSQL 常量与 contracts 源文件逐字节一致（契约唯一真源守护）。
func TestSchemaSQLMatchesContracts(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "packages", "contracts", "src", "schema.sql"))
	if err != nil {
		t.Fatalf("read contracts schema: %v", err)
	}
	if schemaSQL != string(raw) {
		t.Fatal("migrate.go 内嵌 schemaSQL 与 packages/contracts/src/schema.sql 不一致")
	}
}

// 迁移幂等：同一库打开两次不报错（IF NOT EXISTS + 版本行去重）。
func TestMigrateIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tako.db")
	st1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	_ = st1.Close()
	st2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	_ = st2.Close()
}

func TestServicesCRUD(t *testing.T) {
	st := tempStore(t)

	b, err := st.SaveService(Service{
		Name: "B 服务", Type: "deepl", Enabled: true,
		Config: ServiceConfig{APIKey: "k", Order: 2},
	})
	if err != nil {
		t.Fatalf("save B: %v", err)
	}
	a, err := st.SaveService(Service{
		Name: "A 服务", Type: "llm", Enabled: true,
		Config: ServiceConfig{BaseURL: "https://x", Model: "m", Order: 1},
	})
	if err != nil {
		t.Fatalf("save A: %v", err)
	}
	if a.ID == "" || b.ID == "" || a.ID == b.ID {
		t.Fatalf("bad ids: %q %q", a.ID, b.ID)
	}
	if a.CreatedAt == 0 || a.UpdatedAt == 0 {
		t.Fatal("timestamps not set")
	}

	list, err := st.ListServices()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 || list[0].ID != a.ID || list[1].ID != b.ID {
		t.Fatalf("order wrong: %+v", list)
	}
	if list[0].Config.BaseURL != "https://x" || list[0].Config.Order != 1 {
		t.Fatalf("config roundtrip wrong: %+v", list[0].Config)
	}

	// 更新保留 created_at
	a.Config.Model = "m2"
	a2, err := st.SaveService(a)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if a2.CreatedAt != a.CreatedAt || a2.UpdatedAt < a.UpdatedAt {
		t.Fatalf("timestamps: %+v", a2)
	}
	got, err := st.GetService(a.ID)
	if err != nil || got.Config.Model != "m2" {
		t.Fatalf("get: %+v %v", got, err)
	}

	if err := st.DeleteService(a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := st.GetService(a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("after delete: %v", err)
	}
}

func TestHistoryCRUDAndQuery(t *testing.T) {
	st := tempStore(t)

	items, total, err := st.ListHistory(HistoryQuery{})
	if err != nil || total != 0 || len(items) != 0 {
		t.Fatalf("empty list: %v %d %d", err, total, len(items))
	}

	h1, _ := st.SaveHistory(HistoryItem{
		SourceText: "hello world", SourceLang: "en", TargetLang: "zh-Hans",
		ServiceID: "svc-a", ServiceName: "A", ResultText: "你好世界",
	})
	h2, _ := st.SaveHistory(HistoryItem{
		SourceText: "goodbye", SourceLang: "en", TargetLang: "zh-Hans",
		ServiceID: "svc-b", ServiceName: "B", ResultText: "再见",
	})
	if _, err := st.SaveHistory(HistoryItem{
		SourceText: "hello again", SourceLang: "en", TargetLang: "zh-Hans",
		ServiceID: "svc-a", ServiceName: "A", ResultText: "又见面了",
	}); err != nil {
		t.Fatalf("save h3: %v", err)
	}

	// 关键词模糊（中英文分别命中 source/result）
	_, total, _ = st.ListHistory(HistoryQuery{Keyword: "hello"})
	if total != 2 {
		t.Fatalf("keyword hello total=%d, want 2", total)
	}
	_, total, _ = st.ListHistory(HistoryQuery{Keyword: "再见"})
	if total != 1 {
		t.Fatalf("keyword 再见 total=%d, want 1", total)
	}
	// serviceId 过滤
	_, total, _ = st.ListHistory(HistoryQuery{ServiceID: "svc-a"})
	if total != 2 {
		t.Fatalf("serviceId total=%d, want 2", total)
	}
	// 分页
	items, total, _ = st.ListHistory(HistoryQuery{Limit: 2})
	if total != 3 || len(items) != 2 {
		t.Fatalf("page: total=%d len=%d", total, len(items))
	}
	items2, _, _ := st.ListHistory(HistoryQuery{Limit: 2, Offset: 2})
	if len(items2) != 1 {
		t.Fatalf("offset page len=%d", len(items2))
	}

	// 收藏
	if err := st.ToggleFavorite(h2.ID, true); err != nil {
		t.Fatalf("fav: %v", err)
	}
	favItems, total, _ := st.ListHistory(HistoryQuery{FavoriteOnly: true})
	if total != 1 || favItems[0].ID != h2.ID {
		t.Fatalf("favorite list: %+v", favItems)
	}
	if err := st.ToggleFavorite("nonexist", true); !errors.Is(err, ErrNotFound) {
		t.Fatalf("fav nonexist: %v", err)
	}

	if err := st.DeleteHistory(h1.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, total, _ = st.ListHistory(HistoryQuery{})
	if total != 2 {
		t.Fatalf("after delete total=%d", total)
	}
	if err := st.ClearHistory(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	_, total, _ = st.ListHistory(HistoryQuery{})
	if total != 0 {
		t.Fatalf("after clear total=%d", total)
	}
}

func TestSettingsKV(t *testing.T) {
	st := tempStore(t)

	_, found, err := st.GetSetting("app")
	if err != nil || found {
		t.Fatalf("missing key: found=%v err=%v", found, err)
	}

	type cfg struct {
		Theme string `json:"theme"`
	}
	if err := st.SetSetting("app", json.RawMessage(`{"theme":"dark"}`)); err != nil {
		t.Fatalf("set: %v", err)
	}
	raw, found, err := st.GetSetting("app")
	if err != nil || !found {
		t.Fatalf("get: found=%v err=%v", found, err)
	}
	var c cfg
	if err := json.Unmarshal(raw, &c); err != nil || c.Theme != "dark" {
		t.Fatalf("unmarshal: %v %+v", err, c)
	}

	// 覆盖写
	if err := st.SetSetting("app", json.RawMessage(`{"theme":"light"}`)); err != nil {
		t.Fatalf("update: %v", err)
	}
	raw, _, _ = st.GetSetting("app")
	_ = json.Unmarshal(raw, &c)
	if c.Theme != "light" {
		t.Fatalf("theme=%q", c.Theme)
	}
}
