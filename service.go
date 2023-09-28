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
			priority int not null default 1,
			name varchar(255) not null unique
		);

		create table if not exists permissions (
			id serial primary key,
			key varchar(255) not null unique,
			value int not null default 0
		);

		create table if not exists group_permissions (
			group_id int not null, 
			permission_id int not null,
			value int not null,
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

		create or replace function get_group_permissions(int)
		returns setof permissions
		as $$
		begin
			return query
			select p.id, p.key, coalesce(gp.value, p.value)  as value
			from 
				permissions p
			left join 
				group_permissions gp on p.id = gp.permission_id and gp.group_id = $1
			order by p.key asc;
		end;
		$$ language plpgsql;

		create or replace function get_user_permissions(int) 
		returns setof permissions
		as $$
		declare
			v_group groups;
		begin
			select g.* 
			into v_group 
			from 
				groups g
			inner join
				group_users gu on gu.group_id = g.id and gu.user_id = $1
			order by 
				g.priority desc
			limit 1;

			if not found then
				return query select * from permissions order by key asc;
			end if;

			return query
			select p.id, p.key, coalesce(gp.value, p.value)  as value
			from 
				group_permissions gp
			left join 
				permissions p on p.id = gp.permission_id
			where 
				gp.group_id = v_group.id
			order by p.key asc;
		end;
		$$ language plpgsql;
	`)
}
