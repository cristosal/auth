package auth

import (
	"github.com/cristosal/orm"
)

type Service struct {
	db             orm.DB
	permissionRepo *PermissionRepo
	userRepo       *UserRepo
	groupRepo      *GroupRepo
	sessionRepo    *SessionRepo
}

func NewService(db orm.DB) *Service {
	return &Service{
		db:             db,
		permissionRepo: NewPermissionRepo(db),
		groupRepo:      NewGroupRepo(db),
		userRepo:       NewUserRepo(db),
		sessionRepo:    NewSessionRepo(db),
	}
}

func (s *Service) Sessions() *SessionRepo {
	return s.sessionRepo
}

func (s *Service) Users() *UserRepo {
	return s.userRepo
}

func (s *Service) Permissions() *PermissionRepo {
	return s.permissionRepo
}

func (s *Service) Groups() *GroupRepo {
	return s.groupRepo
}

func (s *Service) Init() error {
	return orm.AddMigrations(s.db, migrations)
}
