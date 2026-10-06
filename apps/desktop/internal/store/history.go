// history 表 CRUD + 分页过滤查询（keyword/serviceId/favoriteOnly）。
// 字段对齐 contracts.HistoryItem；时间戳 epoch ms。
package store

import (
	"fmt"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

// HistoryItem 一条翻译历史。
type HistoryItem struct {
	ID          string
	CreatedAt   int64 // epoch ms
	SourceText  string
	SourceLang  string
	TargetLang  string
	ServiceID   string
	ServiceName string // 冗余存储：服务删除后仍可展示
	ResultText  string
	Favorite    bool
}

// HistoryQuery 分页过滤参数（零值即不加该过滤）。
type HistoryQuery struct {
	Keyword      string // 模糊匹配 sourceText/resultText
	ServiceID    string
	FavoriteOnly bool
	Limit        int
	Offset       int
}

// ListHistory 分页查询（created_at 倒序）+ 满足条件的总行数。
func (s *Store) ListHistory(q HistoryQuery) ([]HistoryItem, int, error) {
	where, args := historyWhere(q)

	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM history`+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := max(q.Offset, 0)

	rows, err := s.db.Query(`
		SELECT id, created_at, source_text, source_lang, target_lang,
		       service_id, service_name, result_text, favorite
		FROM history`+where+`
		ORDER BY created_at DESC, id
		LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []HistoryItem{}
	for rows.Next() {
		var it HistoryItem
		var fav int
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.SourceText, &it.SourceLang,
			&it.TargetLang, &it.ServiceID, &it.ServiceName, &it.ResultText, &fav); err != nil {
			return nil, 0, err
		}
		it.Favorite = fav != 0
		items = append(items, it)
	}
	return items, total, rows.Err()
}

// historyWhere 组装 WHERE 子句与参数（keyword 用 LIKE 模糊，escape % _ \）。
func historyWhere(q HistoryQuery) (string, []any) {
	var clauses []string
	var args []any
	if kw := strings.TrimSpace(q.Keyword); kw != "" {
		clauses = append(clauses, `(source_text LIKE ? ESCAPE '\' OR result_text LIKE ? ESCAPE '\')`)
		escaped := escapeLike(kw)
		args = append(args, "%"+escaped+"%", "%"+escaped+"%")
	}
	if q.ServiceID != "" {
		clauses = append(clauses, `service_id = ?`)
		args = append(args, q.ServiceID)
	}
	if q.FavoriteOnly {
		clauses = append(clauses, `favorite = 1`)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return ` WHERE ` + strings.Join(clauses, ` AND `), args
}

// escapeLike 转义 LIKE 通配符（与 ESCAPE '\' 配套）。
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

// SaveHistory 保存一条历史（ID 空则生成；created_at 缺省取当前时间）。
func (s *Store) SaveHistory(it HistoryItem) (HistoryItem, error) {
	if it.ID == "" {
		id, err := gonanoid.New()
		if err != nil {
			return HistoryItem{}, err
		}
		it.ID = id
	}
	if it.CreatedAt == 0 {
		it.CreatedAt = time.Now().UnixMilli()
	}
	fav := 0
	if it.Favorite {
		fav = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO history (id, created_at, source_text, source_lang, target_lang,
		                     service_id, service_name, result_text, favorite)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			source_text = excluded.source_text,
			source_lang = excluded.source_lang,
			target_lang = excluded.target_lang,
			service_id = excluded.service_id,
			service_name = excluded.service_name,
			result_text = excluded.result_text`,
		it.ID, it.CreatedAt, it.SourceText, it.SourceLang, it.TargetLang,
		it.ServiceID, it.ServiceName, it.ResultText, fav)
	if err != nil {
		return HistoryItem{}, err
	}
	return it, nil
}

// DeleteHistory 删除一条历史。
func (s *Store) DeleteHistory(id string) error {
	_, err := s.db.Exec(`DELETE FROM history WHERE id = ?`, id)
	return err
}

// ToggleFavorite 设置收藏状态。
func (s *Store) ToggleFavorite(id string, favorite bool) error {
	fav := 0
	if favorite {
		fav = 1
	}
	res, err := s.db.Exec(`UPDATE history SET favorite = ? WHERE id = ?`, fav, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("history %s: %w", id, ErrNotFound)
	}
	return nil
}

// ClearHistory 清空历史（含收藏）。
func (s *Store) ClearHistory() error {
	_, err := s.db.Exec(`DELETE FROM history`)
	return err
}
