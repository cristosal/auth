package auth

import (
	"github.com/cristosal/orm"
)

type PgxService struct {
	db         orm.DB
	permission *PermissionPgxRepo
	user       *UserPgxService
	group      *GroupPgxRepo
	sessions   *PgxSessionStore
}

func NewPgxService(db orm.DB) *PgxService {
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
	return orm.AddMigrations(s.db, migrations)
}
