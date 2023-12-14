package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/cristosal/orm"
	"github.com/cristosal/orm/schema"
)

type (

	// SessionRepo is a postgres backed session store
	SessionRepo struct{ db orm.DB }

	sessionRow struct {
		ID        string
		UserID    *int64
		Data      Session
		CreatedAt time.Time
		UpdatedAt time.Time
		ExpiresAt time.Time
	}
)

func (sessionRow) TableName() string {
	return "sessions"
}

// NewSessionRepo returns postgres backed session store
func NewSessionRepo(db orm.DB) *SessionRepo {
	return &SessionRepo{db}
}

// Drop drops the session table
func (s *SessionRepo) Drop() error {
	return orm.Exec(s.db, "drop table sessions")
}

// Save upserts session into database
func (s *SessionRepo) Save(sess *Session) error {
	sess.Counter++

	if sess.ID == "" {
		sid, err := GenerateToken(16)
		if err != nil {
			return err
		}

		sess.ID = sid
		return orm.Exec(s.db, "insert into sessions (id, user_id, data, expires_at) values ($1, $2, $3, $4)",
			sid, sess.UserID(), sess, sess.ExpiresAt)
	}

	return orm.Exec(s.db, "update sessions set updated_at = now(), data = $1, user_id = $2 where id = $3", sess, sess.UserID(), sess.ID)
}

// ByID returns a session by its id
func (s *SessionRepo) ByID(sessionID string) (*Session, error) {
	var row sessionRow
	if err := orm.Get(s.db, &row, "where id = $1", sessionID); err != nil {
		if errors.Is(err, orm.ErrNotFound) {
			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	return &row.Data, nil
}

// ByUserID returns all sessions belonging to a user
func (s *SessionRepo) ByUserID(uid int64) ([]Session, error) {
	var rows []sessionRow
	if err := orm.List(s.db, &rows, "user_id = $1", uid); err != nil {
		return nil, err
	}

	sessions := make([]Session, 0)
	for i := range rows {
		sessions = append(sessions, rows[i].Data)
	}

	return sessions, nil
}

// Remove session by id
func (s *SessionRepo) RemoveByID(id string) error {
	return orm.Exec(s.db, "delete from sessions where id = $1", id)
}

// DeleteByUserID deletes all sessions for users in the email list
func (s *SessionRepo) RemoveByEmails(emails []string) error {
	valueList := schema.ValueList(len(emails), 1)
	sql := fmt.Sprintf(`DELETE FROM sessions WHERE user_id IN (SELECT id FROM users WHERE email IN (%s))`, valueList)
	var values []any
	for i := range emails {
		values = append(values, emails[i])
	}

	return orm.Exec(s.db, sql, values...)
}

// RemoveByUserID deletes all sessions for a given user
func (s *SessionRepo) RemoveByUserID(uid int64) error {
	return orm.Exec(s.db, "delete from sessions where user_id = $1", uid)
}

// RemoveExpired deletes all sessions which have expired
func (s *SessionRepo) RemoveExpired() error {
	return orm.Exec(s.db, "delete from sessions where expires_at < now()")
}
