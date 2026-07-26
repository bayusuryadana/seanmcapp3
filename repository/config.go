package repository

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

const (
	ConfigKeyIGSessionID = "ig_session_id"
	ConfigKeyIGCSRFToken = "ig_csrf_token"
)

type ConfigRepo interface {
	GetValues(keys ...string) (map[string]string, error)
	SetValues(values map[string]string) error
}

type ConfigRepoImpl struct {
	DB *sql.DB
}

func (r *ConfigRepoImpl) GetValues(keys ...string) (map[string]string, error) {
	result := make(map[string]string)
	if len(keys) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, key := range keys {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = key
	}

	query := fmt.Sprintf(
		"SELECT key, value FROM app_config WHERE key IN (%s)",
		strings.Join(placeholders, ", "),
	)
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *ConfigRepoImpl) SetValues(values map[string]string) error {
	keys := make([]string, 0, len(values))
	for key, value := range values {
		if key != "" && value != "" {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)

	placeholders := make([]string, len(keys))
	args := make([]any, 0, len(keys)*2)
	for i, key := range keys {
		placeholders[i] = fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2)
		args = append(args, key, values[key])
	}

	query := fmt.Sprintf(
		"INSERT INTO app_config (key, value) VALUES %s ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()",
		strings.Join(placeholders, ", "),
	)
	_, err := r.DB.Exec(query, args...)
	return err
}
