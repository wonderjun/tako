// 迁移器：schema 唯一依据是 packages/contracts/src/schema.sql，此处内嵌其
// 常量副本（go:embed 无法跨 module 边界，contracts 在 module 外）。
// migrate_test.go 守护常量与源文件逐字节一致。
package store

import (
	"database/sql"
	"fmt"
)

// schemaSQL 与 packages/contracts/src/schema.sql 逐字节一致（v1）。
const schemaSQL = `-- ==============================================================================
-- Tako SQLite schema（v1）— internal/store 迁移的唯一依据
-- 库文件位置：用户数据目录（macOS ~/Library/Application Support/Tako/tako.db）
-- 时间戳一律 epoch ms（INTEGER）；服务私密配置整体 JSON 存 config 列。
-- ==============================================================================

PRAGMA journal_mode = WAL;

-- 翻译服务配置
CREATE TABLE IF NOT EXISTS services (
  id         TEXT PRIMARY KEY,           -- nanoid
  name       TEXT NOT NULL,
  type       TEXT NOT NULL,              -- llm | deepl | google | microsoft
  enabled    INTEGER NOT NULL DEFAULT 1,
  config     TEXT NOT NULL,              -- JSON: baseUrl/apiKey/model/authScheme/customHeaders/order
  created_at INTEGER NOT NULL,
  updated_at INTEGER NOT NULL
);

-- 翻译历史
CREATE TABLE IF NOT EXISTS history (
  id           TEXT PRIMARY KEY,         -- nanoid
  created_at   INTEGER NOT NULL,
  source_text  TEXT NOT NULL,
  source_lang  TEXT NOT NULL,            -- 'auto' 或语言代码
  target_lang  TEXT NOT NULL,
  service_id   TEXT NOT NULL,
  service_name TEXT NOT NULL,            -- 冗余存储，服务删除后仍可展示
  result_text  TEXT NOT NULL,
  favorite     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_history_created_at ON history (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_favorite ON history (favorite) WHERE favorite = 1;

-- 应用设置（KV，value 为 JSON）
CREATE TABLE IF NOT EXISTS settings (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);

-- schema 版本（迁移器维护，勿手改）
CREATE TABLE IF NOT EXISTS schema_version (
  version INTEGER NOT NULL
);
`

// schemaVersion 当前 schema 版本（对应 contracts schema.sql 的 v1）。
const schemaVersion = 1

// migrate 执行迁移：建表语句全部 IF NOT EXISTS，天然幂等；随后保证
// schema_version 表存在版本行。跑两次不报错是验收硬指标。
func migrate(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return fmt.Errorf("执行 schema 失败: %w", err)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_version`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if _, err := db.Exec(`INSERT INTO schema_version (version) VALUES (?)`, schemaVersion); err != nil {
			return err
		}
	}
	return nil
}
