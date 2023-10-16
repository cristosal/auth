package auth

import "github.com/cristosal/pgxx"

// GroupPermission represents the union between a group and a permission
// it can contain a value for use in application
type GroupPermission struct {
	GroupID      pgxx.ID
	PermissionID pgxx.ID
	Priority     int    `db:"-"`
	Name         string `db:"-"`
	Value        int
}

func (*GroupPermission) TableName() string {
	return "group_permissions"
}

type GroupPermissions []GroupPermission

// Value returns the value associated with the permission of a given name.
// it takes into account conflicting permissions and takes the one with higher priority
func (gps GroupPermissions) Value(name string) int {
	var (
		priority *int
		value    = 0
	)

	for i := range gps {
		if gps[i].Name == name && (priority == nil || gps[i].Priority > *priority) {
			priority = &gps[i].Priority
			value = gps[i].Value
		}
	}

	return value
}

func (gps GroupPermissions) Has(name string) bool {
	for i := range gps {
		if gps[i].Name == name {
			return true
		}
	}

	return false
}

func (s *Service) UserGroupPermissions(uid pgxx.ID) (GroupPermissions, error) {
	sql := `select 
		gp.group_id, 
		gp.permission_id, 
		g.priority,
		p.name,
		gp.value
	from 
		group_permissions gp 
	inner join 
		permissions p 
	on 
		p.id = gp.permission_id 
	inner join
		groups g
	on
		g.id = gp.group_id
	inner join
		group_users gu
	on
		gu.group_id = g.id
	where 
		gu.user_id = $1`

	rows, err := s.db.Query(ctx, sql, uid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	groupPermissions := make([]GroupPermission, 0)

	for rows.Next() {
		var gp GroupPermission
		err := rows.Scan(
			&gp.GroupID,
			&gp.PermissionID,
			&gp.Priority,
			&gp.Name,
			&gp.Value,
		)

		if err != nil {
			return nil, err
		}

		groupPermissions = append(groupPermissions, gp)
	}

	return groupPermissions, nil
}

// GroupPermissions returns group permissions for a group by group id
func (s *Service) GroupPermissions(gid pgxx.ID) (GroupPermissions, error) {
	sql := `select 
		gp.group_id, 
		gp.permission_id, 
		g.priority,
		p.name,
		gp.value
	from 
		group_permissions gp 
	inner join 
		permissions p 
	on 
		p.id = gp.permission_id 
	inner join
		groups g
	on
		g.id = gp.group_id
	where 
		gp.group_id = $1`

	rows, err := s.db.Query(ctx, sql, gid)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	groupPermissions := make([]GroupPermission, 0)

	for rows.Next() {
		var gp GroupPermission
		err := rows.Scan(
			&gp.GroupID,
			&gp.PermissionID,
			&gp.Priority,
			&gp.Name,
			&gp.Value,
		)

		if err != nil {
			return nil, err
		}

		groupPermissions = append(groupPermissions, gp)
	}

	return groupPermissions, nil
}

func (s *Service) AddGroupPermission(gid, pid pgxx.ID, v int) error {
	return pgxx.Exec(s.db, "insert into group_permissions (group_id, permission_id, value) values ($1, $2, $3)", gid, pid, v)
}

func (s *Service) RemoveGroupPermission(gid, pid pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from group_permissions where group_id = $1 and permission_id = $2", gid, pid)
}
