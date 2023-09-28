package auth

import (
	"fmt"
	"strings"

	"github.com/cristosal/pgxx"
)

type (
	// Permission represents a users access to a resource.
	// A permission is only granted when the value > 1
	Permission struct {
		ID    pgxx.ID
		Key   string // Unique name for permission
		Value int    // value associated with permission in case
	}

	// Permissions is the collection type of Permission
	Permissions []Permission
)

// Has returns true when a permission with a given key is found and it's value is greater than 0
func (p Permissions) Has(key string) bool {
	for i := range p {
		if p[i].Key == key {
			if p[i].Value > 0 {
				return true
			}
		}
	}
	return false
}

// Value returns the value associated with a permission
func (p Permissions) Value(key string) int {
	for _, perm := range p {
		if perm.Key == key {
			return perm.Value
		}
	}
	return 0
}

func (p *Permission) TableName() string {
	return "permissions"
}

func (s *Service) UpdatePermission(p *Permission) error {
	return pgxx.Update(s.db, p)
}

func (s *Service) Permissions() (Permissions, error) {
	var perms []Permission
	err := pgxx.Many(s.db, &perms, "order by key asc")
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *Service) PermissionsByGroup(gid pgxx.ID) (Permissions, error) {
	rows, err := s.db.Query(ctx, `select * from get_group_permissions($1)`, gid)
	if err != nil {
		return nil, err
	}

	return pgxx.CollectRows[Permission](rows)
}

func (s *Service) PermissionsByUser(uid pgxx.ID) (Permissions, error) {
	rows, err := s.db.Query(ctx, `select * from get_user_permissions($1)`, uid)

	if err != nil {
		return nil, err
	}

	return pgxx.CollectRows[Permission](rows)
}

func (s *Service) AssignPermissions(gid pgxx.ID, permissions Permissions) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := pgxx.Exec(tx, "delete from group_permissions where group_id = $1", gid); err != nil {
		return err
	}

	if len(permissions) == 0 {
		return tx.Commit(ctx)
	}

	valuesql := []string{}
	var values []any
	for i, p := range permissions {
		valuesql = append(valuesql, fmt.Sprintf("(%d, %d, $%d)", gid, p.ID, i+1))
		values = append(values, p.Value)
	}

	sql := fmt.Sprintf("insert into group_permissions (group_id, permission_id, value) values %s", strings.Join(valuesql, ", "))
	if err := pgxx.Exec(tx, sql, values...); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Service) CreatePermission(key string, v int) error {
	return pgxx.Exec(s.db, `insert into permissions (key, value) values ($1, $2)`, key, v)
}

func (s *Service) DeletePermission(id pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from permissions where id = $1", id)
}
