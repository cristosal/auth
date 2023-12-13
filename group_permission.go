package auth

import (
	"github.com/cristosal/orm"
)

// GroupPermission represents the union between a group and a permission
// it can contains a value for use in application logic
type GroupPermission struct {
	GroupID      int64
	PermissionID int64
	Priority     int    `db:"-"` // group priority value
	Name         string `db:"-"` // permission name
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

func (r *GroupRepo) UserPermissions(uid int64) (GroupPermissions, error) {
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

	rows, err := r.db.Query(sql, uid)
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

// Permissions returns group permissions for a group by group id
func (r *GroupRepo) Permissions(gid int64) (GroupPermissions, error) {
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

	rows, err := r.db.Query(sql, gid)
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

func (r *GroupRepo) AddPermission(gid, pid int64, value int) error {
	return orm.Exec(r.db, "insert into group_permissions (group_id, permission_id, value) values ($1, $2, $3) on conflict do nothing", gid, pid, value)
}

func (r *GroupRepo) RemovePermission(gid, pid int64) error {
	return orm.Exec(r.db, "delete from group_permissions where group_id = $1 and permission_id = $2", gid, pid)
}
