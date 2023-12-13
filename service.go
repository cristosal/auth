package auth

import (
	"github.com/cristosal/orm"
)

type Service struct {
	db         orm.DB
	permission *PermissionRepo
	user       *UserService
	group      *GroupRepo
	sessions   *SessionStore
}

func NewService(db orm.DB) *Service {
	return &Service{
		db:         db,
		permission: NewPermissionRepo(db),
		group:      NewGroupRepo(db),
		user:       NewUserService(db),
		sessions:   NewSessionStore(db),
	}
}

func (s *Service) Sessions() *SessionStore {
	return s.sessions
}

func (s *Service) Users() *UserService {
	return s.user
}

func (s *Service) Permissions() *PermissionRepo {
	return s.permission
}

func (s *Service) Groups() *GroupRepo {
	return s.group
}

func (s *Service) Init() error {
	return orm.AddMigrations(s.db, migrations)
}
