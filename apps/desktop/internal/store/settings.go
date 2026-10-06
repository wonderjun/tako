// settings 表：KV 存取，value 为 JSON 文本（AppSettings 等整体序列化）。
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
)

// GetSetting 读取 key 的 JSON 值；不存在返回 found=false。
func (s *Store) GetSetting(key string) (json.RawMessage, bool, error) {
	var raw string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return json.RawMessage(raw), true, nil
}

// SetSetting 写入 key 的 JSON 值（整体覆盖）。
func (s *Store) SetSetting(key string, value json.RawMessage) error {
	_, err := s.db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, string(value))
	return err
}
