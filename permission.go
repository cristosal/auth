package auth

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cristosal/pgxx"
)

type PermissionType string

const (
	Quantity PermissionType = "quantity"
	Access   PermissionType = "access"
)

// Permission represents a users access to a resource.
type Permission struct {
	ID          pgxx.ID
	Name        string
	Description string
	Type        PermissionType
}

func (p *Permission) TableName() string {
	return "permissions"
}

// Permissions is a collection of permission
type Permissions []Permission

// Has returns true when a permission with a given key is found and it's value is greater than 0
func (p Permissions) Has(name string) bool {
	for i := range p {
		if p[i].Name == name {
			return true
		}
	}
	return false
}

type PermissionPgxRepo struct{ db pgxx.DB }

func NewPermissionPgxRepo(db pgxx.DB) *PermissionPgxRepo {
	return &PermissionPgxRepo{db}
}

func (r *PermissionPgxRepo) Seed(permissions []Permission) error {
	var (
		i     = 1
		parts []string
		args  []any
	)

	for _, v := range permissions {
		parts = append(parts, fmt.Sprintf("($%d, $%d, $%d)", i, i+1, i+2))
		args = append(args, v.Name, v.Description, v.Type)
		i += 3
	}

	sql := fmt.Sprintf("insert into permissions (name, description, type) values %s on conflict (name) do nothing", strings.Join(parts, ", "))
	if err := pgxx.Exec(r.db, sql, args...); err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	for i := range permissions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			pgxx.One(r.db, &permissions[i], "where name = $1", permissions[i].Name)
		}(i)
	}

	wg.Wait()
	return nil
}

// List lists all permissions
func (s *PermissionPgxRepo) List() (Permissions, error) {
	var perms []Permission
	err := pgxx.Many(s.db, &perms, "order by name asc")
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *PermissionPgxRepo) Add(p *Permission) error {
	return pgxx.Insert(s.db, p)
}

func (s *PermissionPgxRepo) Update(p *Permission) error {
	return pgxx.Update(s.db, p)
}

func (s *PermissionPgxRepo) Clear() error {
	return pgxx.Exec(s.db, "delete from permissions")
}

func (s *PermissionPgxRepo) Remove(id pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from permissions where id = $1", id)
}
