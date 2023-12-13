package auth

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cristosal/orm"
)

type PermissionType string

const (
	Quantity PermissionType = "quantity"
	Access   PermissionType = "access"
)

// Permission represents a users access to a resource.
type Permission struct {
	ID          int64
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

type PermissionRepo struct{ db orm.DB }

func NewPermissionRepo(db orm.DB) *PermissionRepo {
	return &PermissionRepo{db}
}

func (r *PermissionRepo) Seed(permissions []Permission) error {
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
	if err := orm.Exec(r.db, sql, args...); err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	for i := range permissions {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			orm.Get(r.db, &permissions[i], "where name = $1", permissions[i].Name)
		}(i)
	}

	wg.Wait()
	return nil
}

// List lists all permissions
func (s *PermissionRepo) List() (Permissions, error) {
	var perms []Permission
	err := orm.List(s.db, &perms, "order by name asc")
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *PermissionRepo) Add(p *Permission) error {
	return orm.Add(s.db, p)
}

func (s *PermissionRepo) Update(p *Permission) error {
	return orm.UpdateByID(s.db, p)
}

func (s *PermissionRepo) Clear() error {
	return orm.Exec(s.db, "delete from permissions")
}

func (s *PermissionRepo) Remove(id int64) error {
	return orm.Exec(s.db, "delete from permissions where id = $1", id)
}

func (s *PermissionRepo) RemoveByName(name string) error {
	return orm.Exec(s.db, "delete from permissions where name = $1", name)
}
