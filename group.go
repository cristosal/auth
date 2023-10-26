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

// GroupPgxRepo us a group repository using pgx
type GroupPgxRepo struct{ db pgxx.DB }

func NewGroupPgxRepo(db pgxx.DB) *GroupPgxRepo {
	return &GroupPgxRepo{db}
}

// Seed seeds groups to the database.
// If they already exist it will not return an error
func (r *GroupPgxRepo) Seed(groups []Group) error {
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
	rows, err := r.db.Query(ctx, sql, args...)
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

// AddUser adds a user to a group.
// No error will occur if a user is already part of the group
func (r *GroupPgxRepo) AddUser(uid pgxx.ID, gid pgxx.ID) error {
	return pgxx.Exec(r.db, "insert into group_users (user_id, group_id) values ($1, $2) on conflict do nothing", uid, gid)
}

// RemoveUser removes a user from a group
func (r *GroupPgxRepo) RemoveUser(uid pgxx.ID, gid pgxx.ID) error {
	return pgxx.Exec(r.db, "delete from group_users where user_id = $1 and group_id = $2", uid, gid)
}

// GroupByName finds a group by it's name
func (r *GroupPgxRepo) ByName(name string) (*Group, error) {
	var g Group
	if err := pgxx.One(r.db, &g, "where name = $1", name); err != nil {
		return nil, err
	}
	return &g, nil
}

// Remove deletes a group by id
func (r *GroupPgxRepo) Remove(gid pgxx.ID) error {
	return pgxx.Exec(r.db, "delete from groups where id = $1", gid)
}

// GroupsByUser returns all groups that user is a part of
func (r *GroupPgxRepo) ByUser(uid pgxx.ID) (Groups, error) {
	var groups []Group
	if err := pgxx.Many(r.db, &groups, "inner join group_users gu on gu.user_id = $1", uid); err != nil {
		return nil, err
	}
	return groups, nil
}

// Groups returns all groups ordered by priority (highest first)
func (r *GroupPgxRepo) List() (Groups, error) {
	var groups []Group
	err := pgxx.Many(r.db, &groups, "order by priority desc")
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GroupByID returns a group by it's id
func (r *GroupPgxRepo) ByID(id pgxx.ID) (*Group, error) {
	var g Group
	if err := pgxx.One(r.db, &g, "where id = $1", id); err != nil {
		return nil, err
	}
	return &g, nil
}

// Add adds a group
func (r *GroupPgxRepo) Add(g *Group) error {
	return pgxx.Insert(r.db, g)
}

// Update updates a group with name and priority
func (r *GroupPgxRepo) Update(g *Group) error {
	return pgxx.Update(r.db, g)
}

// GroupUserCount counts all users within a group
func (r *GroupPgxRepo) UserCount(gid pgxx.ID) (int, error) {
	row := r.db.QueryRow(ctx, "select count(*) from group_users where group_id = $1", gid)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GroupUsers returns all users within a group
func (r *GroupPgxRepo) Users(gid pgxx.ID) ([]User, error) {
	cols := pgxx.MustAnalyze(&User{}).Fields.Columns().PrefixedList("u")
	sql := fmt.Sprintf("select %s from users u inner join group_users gu on u.id = gu.user_id where gu.group_id = $1 order by u.created_at desc", cols)
	rows, err := r.db.Query(ctx, sql, gid)
	if err != nil {
		return nil, err
	}

	return pgxx.CollectRows[User](rows)
}
