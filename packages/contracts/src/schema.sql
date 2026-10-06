-- ==============================================================================
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
