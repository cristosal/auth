package auth

import (
	"context"

	"github.com/cristosal/pgxx"
)

type PgxService struct {
	db         pgxx.DB
	permission *PermissionPgxRepo
	user       *UserPgxService
	group      *GroupPgxRepo
	sessions   *PgxSessionStore
}

var ctx = context.Background()

func NewPgxService(db pgxx.DB) *PgxService {
	return &PgxService{
		db:         db,
		permission: NewPermissionPgxRepo(db),
		group:      NewGroupPgxRepo(db),
		user:       NewUserPgxService(db),
		sessions:   NewPgxSessionStore(db),
	}
}

func (s *PgxService) Sessions() *PgxSessionStore {
	return s.sessions
}

func (s *PgxService) Users() *UserPgxService {
	return s.user
}

func (s *PgxService) Permissions() *PermissionPgxRepo {
	return s.permission
}

func (s *PgxService) Groups() *GroupPgxRepo {
	return s.group
}

func (s *PgxService) Init() error {
	return pgxx.Exec(s.db, `
		create table if not exists users (
			id serial primary key,
			name varchar(255) not null,
			email varchar(1024) not null unique,
			phone varchar(255) not null,
			password varchar(1024) not null,
			confirmed_at timestamptz,
			last_login timestamptz,
			created_at timestamptz not null default current_timestamp,
			updated_at timestamptz not null default current_timestamp
		);

		create table if not exists sessions (
			id varchar(64) primary key not null,
			user_id int,
			data jsonb not null,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now(),
			expires_at timestamptz not null,
			foreign key (user_id) references users(id)
		);

		create table if not exists pass_tokens (
			user_id int not null references users (id) on delete cascade,
			token varchar(64) not null,
			expires timestamptz,
			primary key (user_id)
		);

		create table if not exists registration_tokens (
			user_id int not null references users (id) on delete cascade,
			token varchar(64) not null,
			expires timestamptz not null,
			primary key (user_id)
		);

		create table if not exists groups (
			id serial primary key,
			name varchar(255) not null unique,
			description text not null,
			priority int not null default 1
		);

		create table if not exists permissions (
			id serial primary key,
			name varchar(255) not null unique,
			description text not null,
			type varchar(32) not null default 'access'
		);

		create table if not exists group_permissions (
			group_id int not null, 
			permission_id int not null,
			value int not null default 0,
			primary key (group_id, permission_id),
			foreign key (group_id) references groups (id) on delete cascade,
			foreign key (permission_id) references permissions (id) on delete cascade
		);

		create table if not exists group_users (
			user_id int not null,
			group_id int not null,
			primary key(user_id, group_id),
			foreign key (user_id) references users (id) on delete cascade,
			foreign key (group_id) references groups (id) on delete cascade
		);
	`)
}
