package auth

import (
	"fmt"
	"strings"

	"github.com/cristosal/pgxx"
)

type (
	Group struct {
		ID          pgxx.ID
		Name        string
		Description string
		Priority    int
	}

	Groups []Group
)

func (*Group) TableName() string {
	return "groups"
}

func (g *Group) NewPermission(p *Permission, v int) *GroupPermission {
	return &GroupPermission{
		GroupID:      g.ID,
		PermissionID: p.ID,
		Name:         p.Name,
		Value:        v,
	}
}

func (s *Service) SeedGroups(groups []Group) error {
	var (
		i     = 1
		parts []string
		args  []any
	)

	for _, v := range groups {
		parts = append(parts, fmt.Sprintf("($%d, $%d, $%d)", i, i+1, i+2))
		args = append(args, v.Name, v.Description, v.Priority)
		i += 3
	}

	sql := fmt.Sprintf("insert into groups (name, description, priority) values %s on conflict (name) do nothing returning id", strings.Join(parts, ", "))
	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return err
	}

	defer rows.Close()

	i = 0
	for rows.Next() {
		if err := rows.Scan(&groups[i].ID); err != nil {
			return err
		}
		i++
	}

	return rows.Err()
}

// JoinGroup adds a user to a group.
// No error will occur if a user is already part of the group
func (s *Service) JoinGroup(uid pgxx.ID, gid pgxx.ID) error {
	return pgxx.Exec(s.db, "insert into group_users (user_id, group_id) values ($1, $2) on conflict do nothing", uid, gid)
}

// LeaveGroup removes a user from a group
func (s *Service) LeaveGroup(uid pgxx.ID, gid pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from group_users where user_id = $1 and group_id = $2", uid, gid)
}

// GroupByName finds a group by it's name
func (s *Service) GroupByName(name string) (*Group, error) {
	var g Group
	if err := pgxx.One(s.db, &g, "where name = $1", name); err != nil {
		return nil, err
	}
	return &g, nil
}

// DeleteGroup deletes a group by id
func (s *Service) DeleteGroup(gid pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from groups where id = $1", gid)
}

// GroupsByUser returns all groups that user is a part of
func (s *Service) GroupsByUser(uid pgxx.ID) (Groups, error) {
	var groups []Group
	if err := pgxx.Many(s.db, &groups, "inner join group_users gu on gu.user_id = $1", uid); err != nil {
		return nil, err
	}
	return groups, nil
}

// Groups returns all groups ordered by priority (highest first)
func (s *Service) Groups() (Groups, error) {
	var groups []Group
	err := pgxx.Many(s.db, &groups, "order by priority desc")
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GroupByID returns a group by it's id
func (s *Service) GroupByID(id pgxx.ID) (*Group, error) {
	var g Group
	if err := pgxx.One(s.db, &g, "where id = $1", id); err != nil {
		return nil, err
	}
	return &g, nil
}

// CreateGroup creates a group with permissions
func (s *Service) CreateGroup(g *Group) error {
	return pgxx.Insert(s.db, g)
}

// UpdateGroup updates a group with name and priority
func (s *Service) UpdateGroup(g *Group) error {
	return pgxx.Update(s.db, g)
}

// GroupUserCount counts all users within a group
func (s *Service) GroupUserCount(gid pgxx.ID) (int, error) {
	row := s.db.QueryRow(ctx, "select count(*) from group_users where group_id = $1", gid)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GroupUsers returns all users within a group
func (s *Service) GroupUsers(gid pgxx.ID) ([]User, error) {
	cols := pgxx.MustAnalyze(&User{}).Fields.Columns().PrefixedList("u")
	sql := fmt.Sprintf("select %s from users u inner join group_users gu on u.id = gu.user_id where gu.group_id = $1 order by u.created_at desc", cols)
	rows, err := s.db.Query(ctx, sql, gid)
	if err != nil {
		return nil, err
	}

	return pgxx.CollectRows[User](rows)
}

// AssignGroups assigns groups to a user
func (s *Service) AssignGroups(uid pgxx.ID, gids []pgxx.ID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if err := pgxx.Exec(tx, "delete from group_users where user_id = $1", uid); err != nil {
		return err
	}

	if len(gids) == 0 {
		return tx.Commit(ctx)
	}

	var values []string
	for _, gid := range gids {
		values = append(values, fmt.Sprintf("(%d, %d)", gid, uid))
	}

	sql := fmt.Sprintf("insert into group_users (group_id, user_id) values %s", strings.Join(values, ", "))
	if err := pgxx.Exec(tx, sql); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
