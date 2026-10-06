// services 表 CRUD：行 = 平铺列（id/name/type/enabled/created_at/updated_at）
// + config JSON（baseUrl/apiKey/model/authScheme/customHeaders/order）。
// 私密配置（apiKey 等）只落本地库，整体随 config JSON 存取。
package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// ServiceConfig services.config 列的 JSON 形态（字段名对齐 contracts
// ServiceSetting 的 camelCase；order 按 schema 约定存于 config 内）。
type ServiceConfig struct {
	BaseURL       string            `json:"baseUrl,omitempty"`
	APIKey        string            `json:"apiKey,omitempty"`
	Model         string            `json:"model,omitempty"`
	AuthScheme    string            `json:"authScheme,omitempty"`
	CustomHeaders map[string]string `json:"customHeaders,omitempty"`
	Order         int               `json:"order"`
}

// Service 一条服务配置（config JSON 已解出）。
type Service struct {
	ID        string
	Name      string
	Type      string // llm | deepl | google | microsoft
	Enabled   bool
	Config    ServiceConfig
	CreatedAt int64 // epoch ms
	UpdatedAt int64
}

// ErrNotFound 行不存在（GetService/读回等路径用）。
var ErrNotFound = errors.New("record not found")

// ListServices 全部服务（按 config.order 升序，order 相同按创建时间）。
func (s *Store) ListServices() ([]Service, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, enabled, config, created_at, updated_at
		FROM services
		ORDER BY CAST(json_extract(config, '$.order') AS INTEGER), created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Service
	for rows.Next() {
		svc, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, svc)
	}
	return out, rows.Err()
}

type rowScanner interface{ Scan(dest ...any) error }

func scanService(rs rowScanner) (Service, error) {
	var svc Service
	var enabled int
	var configRaw string
	if err := rs.Scan(&svc.ID, &svc.Name, &svc.Type, &enabled, &configRaw, &svc.CreatedAt, &svc.UpdatedAt); err != nil {
		return Service{}, err
	}
	svc.Enabled = enabled != 0
	if err := json.Unmarshal([]byte(configRaw), &svc.Config); err != nil {
		return Service{}, err
	}
	return svc, nil
}

// SaveService 新增或更新服务：ID 空则生成 nanoid；已存在时保留
// created_at、刷新 updated_at。返回落库后的完整行。
func (s *Store) SaveService(svc Service) (Service, error) {
	if svc.ID == "" {
		id, err := gonanoid.New()
		if err != nil {
			return Service{}, err
		}
		svc.ID = id
	}
	now := time.Now().UnixMilli()
	configRaw, err := json.Marshal(svc.Config)
	if err != nil {
		return Service{}, err
	}
	enabled := 0
	if svc.Enabled {
		enabled = 1
	}
	// upsert：insert 用调用方 created_at（为 0 取 now），update 不动 created_at。
	createdAt := svc.CreatedAt
	if createdAt == 0 {
		createdAt = now
	}
	_, err = s.db.Exec(`
		INSERT INTO services (id, name, type, enabled, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			type = excluded.type,
			enabled = excluded.enabled,
			config = excluded.config,
			updated_at = excluded.updated_at`,
		svc.ID, svc.Name, svc.Type, enabled, string(configRaw), createdAt, now)
	if err != nil {
		return Service{}, err
	}
	return s.GetService(svc.ID)
}

// GetService 按 id 取单条服务。
func (s *Store) GetService(id string) (Service, error) {
	row := s.db.QueryRow(`
		SELECT id, name, type, enabled, config, created_at, updated_at
		FROM services WHERE id = ?`, id)
	svc, err := scanService(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Service{}, ErrNotFound
	}
	return svc, err
}

// DeleteService 删除服务。
func (s *Store) DeleteService(id string) error {
	_, err := s.db.Exec(`DELETE FROM services WHERE id = ?`, id)
	return err
}
