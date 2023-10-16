package auth

import (
	"github.com/cristosal/pgxx"
)

// Permission represents a users access to a resource.
type Permission struct {
	ID          pgxx.ID
	Name        string
	Description string
}

func (p *Permission) TableName() string {
	return "permission"
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

// Permissions lists all permissions
func (s *Service) Permissions() (Permissions, error) {
	var perms []Permission
	err := pgxx.Many(s.db, &perms, "order by name asc")
	if err != nil {
		return nil, err
	}
	return perms, nil
}

func (s *Service) CreatePermission(p *Permission) error {
	return pgxx.Insert(s.db, p)
}

func (s *Service) UpdatePermission(p *Permission) error {
	return pgxx.Update(s.db, p)
}

func (s *Service) DeletePermission(id pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from permissions where id = $1", id)
}
