package auth

import "github.com/cristosal/orm"

var migrations = []orm.Migration{
	{
		Name:        "users table",
		Description: "create users table",
		Up: `create table if not exists users (
				id serial primary key,
				name varchar(255) not null,
				email varchar(1024) not null unique,
				phone varchar(255) not null,
				password varchar(1024) not null,
				confirmed_at timestamptz,
				last_login timestamptz,
				created_at timestamptz not null default current_timestamp,
				updated_at timestamptz not null default current_timestamp
			);`,
		Down: "DROP TABLE users",
	},
	{
		Name:        "sessions table",
		Description: "create sessions table",
		Up: ` create table if not exists sessions (
				id varchar(64) primary key not null,
				user_id int,
				data jsonb not null,
				created_at timestamptz not null default now(),
				updated_at timestamptz not null default now(),
				expires_at timestamptz not null,
				foreign key (user_id) references users(id)
			);`,
		Down: "DROP TABLE sessions",
	},
	{
		Name:        "password tokens table",
		Description: "create password tokens table",
		Up: `create table if not exists pass_tokens (
				user_id int not null references users (id) on delete cascade,
				token varchar(64) not null,
				email varchar(255) not null,
				expires timestamptz,
				primary key (user_id)
			);`,
		Down: "DROP TABLE pass_tokens",
	},
	{
		Name:        "registration tokens table",
		Description: "create registration tokens table",
		Up: `create table if not exists registration_tokens (
				user_id int not null references users (id) on delete cascade,
				email varchar(255) not null,
				token varchar(64) not null,
				expires timestamptz not null,
				primary key (user_id)
			);`,
		Down: "DROP TABLE registration_tokens",
	},

	{
		Name:        "groups table",
		Description: "create groups table",
		Up: `create table if not exists groups (
				id serial primary key,
				name varchar(255) not null unique,
				description text not null,
				priority int not null default 1
			);`,
		Down: "DROP TABLE groups",
	},
	{
		Name:        "permissions table",
		Description: "create permissions table",
		Up: `create table if not exists permissions (
				id serial primary key,
				name varchar(255) not null unique,
				description text not null,
				type varchar(32) not null default 'access'
			);`,
		Down: "DROP TABLE permissions",
	},
	{
		Name:        "group permissions table",
		Description: "create group permissions table",
		Up: `create table if not exists group_permissions (
				group_id int not null, 
				permission_id int not null,
				value int not null default 0,
				primary key (group_id, permission_id),
				foreign key (group_id) references groups (id) on delete cascade,
				foreign key (permission_id) references permissions (id) on delete cascade
			);`,
		Down: "DROP TABLE group_permissions",
	},
	{
		Name:        "group users table",
		Description: "create group users table",
		Up: `create table if not exists group_permissions (
				user_id int not null,
				group_id int not null,
				primary key(user_id, group_id),
				foreign key (user_id) references users (id) on delete cascade,
				foreign key (group_id) references groups (id) on delete cascade
			);`,
		Down: "DROP TABLE group_users",
	},
}
