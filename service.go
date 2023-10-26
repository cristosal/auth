package auth

import (
	"context"

	"github.com/cristosal/pgxx"
)

type Service struct{ db pgxx.DB }

var ctx = context.Background()

func New(db pgxx.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Init() error {
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
			session_id varchar(64) not null primary key,
			user_id int,
			user_agent varchar(1024),
			expires_at timestamptz not null,
			created_at timestamptz not null default current_timestamp,
			foreign key (user_id) references users(id) on delete cascade
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
