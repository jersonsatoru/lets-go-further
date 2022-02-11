package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permission struct {
	ID   int64
	Code string
}

type PermissionModel struct {
	DB *sql.DB
}

type Permissions []string

func (p Permissions) Include(code string) bool {
	for i := range p {
		if p[i] == code {
			return true
		}
	}
	return false
}

func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	query := `
		SELECT p.code
		FROM users u 
			INNER JOIN users_permissions up ON u.id = up.user_id
			INNER JOIN permissions p ON p.id = up.permission_id
		WHERE up.user_id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	permissions := Permissions{}
	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, permissions ...string) error {
	query := `
		INSERT INTO users_permissions 
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)
	`
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	r, err := m.DB.ExecContext(ctx, query, userID, pq.Array(permissions))
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows <= 0 {
		return ErrRecordNotFound
	}
	return nil
}
