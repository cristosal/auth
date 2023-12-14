package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/cristosal/orm"
	"github.com/cristosal/orm/schema"
)

type (
	Group struct {
		ID          int64
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

// GroupRepo us a group repository using pgx
type GroupRepo struct{ db orm.DB }

func NewGroupRepo(db orm.DB) *GroupRepo {
	return &GroupRepo{db}
}

// Seed seeds groups to the database.
// If they already exist it will not return an error
func (r *GroupRepo) Seed(groups []Group) error {
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

	sql := fmt.Sprintf("insert into groups (name, description, priority) values %s on conflict (name) do nothing", strings.Join(parts, ", "))
	err := orm.Exec(r.db, sql, args...)
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	for i := range groups {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			orm.Get(r.db, &groups[i], "where name = $1", groups[i].Name)
		}(i)
	}

	wg.Wait()
	return nil
}

// AddUser adds a user to a group.
// No error will occur if a user is already part of the group
func (r *GroupRepo) AddUser(uid int64, gid int64) error {
	return orm.Exec(r.db, "insert into group_users (user_id, group_id) values ($1, $2) on conflict do nothing", uid, gid)
}

// RemoveUser removes a user from a group
func (r *GroupRepo) RemoveUser(uid int64, gid int64) error {
	return orm.Exec(r.db, "delete from group_users where user_id = $1 and group_id = $2", uid, gid)
}

// GroupByName finds a group by it's name
func (r *GroupRepo) ByName(name string) (*Group, error) {
	var g Group
	if err := orm.Get(r.db, &g, "where name = $1", name); err != nil {
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrGroupNotFound
		}

		return nil, err
	}
	return &g, nil
}

// Remove deletes a group by id
func (r *GroupRepo) Remove(gid int64) error {
	err := orm.Exec(r.db, "delete from groups where id = $1", gid)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrGroupNotFound
	}

	return err
}

// GroupsByUser returns all groups that user is a part of
func (r *GroupRepo) ByUser(uid int64) (Groups, error) {
	var groups []Group
	if err := orm.List(r.db, &groups, "inner join group_users gu on gu.user_id = $1", uid); err != nil {
		return nil, err
	}

	return groups, nil
}

// Groups returns all groups ordered by priority (highest first)
func (r *GroupRepo) List() (Groups, error) {
	var groups []Group
	err := orm.List(r.db, &groups, "order by priority desc")
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GroupByID returns a group by it's id
func (r *GroupRepo) ByID(id int64) (*Group, error) {
	var g Group
	if err := orm.Get(r.db, &g, "where id = $1", id); err != nil {
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrGroupNotFound
		}

		return nil, err
	}
	return &g, nil
}

// Add adds a group
func (r *GroupRepo) Add(g *Group) error {
	if g.Name == "" {
		return ErrNameRequired
	}

	return orm.Add(r.db, g)
}

// Update updates a group with name and priority
func (r *GroupRepo) Update(g *Group) error {
	if g.Name == "" {
		return ErrNameRequired
	}

	return orm.UpdateByID(r.db, g)
}

// GroupUserCount counts all users within a group
func (r *GroupRepo) UserCount(gid int64) (int, error) {
	row := r.db.QueryRow("select count(*) from group_users where group_id = $1", gid)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GroupUsers returns all users within a group
func (r *GroupRepo) Users(gid int64) ([]User, error) {
	cols := schema.MustGet(&User{}).Fields.Columns().PrefixedList("u")
	sql := fmt.Sprintf("select %s from users u inner join group_users gu on u.id = gu.user_id where gu.group_id = $1 order by u.created_at desc", cols)
	rows, err := r.db.Query(sql, gid)
	if err != nil {
		return nil, err
	}

	return orm.CollectRows[User](rows)
}
