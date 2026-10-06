// Package store 提供 SQLite 持久化（services/history/settings 三域）。
// schema 唯一依据是 packages/contracts/src/schema.sql（migrate.go 内嵌其
// 常量副本，migrate_test.go 守护两者一致）。
package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store 数据库句柄。所有方法并发安全（单连接串行化，sqlite 本就单写者）。
type Store struct {
	db *sql.DB
}

// Open 打开（必要时创建）path 处的数据库并完成迁移。
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}
	// busy_timeout 兜底并发写短暂互斥；WAL 模式由 schema 内 PRAGMA 设置。
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// sqlite 单写者：单连接串行化所有访问，根除 SQLITE_BUSY 自锁。
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("schema 迁移失败: %w", err)
	}
	return &Store{db: db}, nil
}

// OpenDefault 打开默认库：os.UserConfigDir()/Tako/tako.db
// （macOS → ~/Library/Application Support/Tako/tako.db）。
func OpenDefault() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("解析用户配置目录失败: %w", err)
	}
	return Open(filepath.Join(dir, "Tako", "tako.db"))
}

// Close 关闭数据库连接。
func (s *Store) Close() error { return s.db.Close() }
