package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cristosal/pgxx"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type (
	// SessionStore is the interface implemented by all stores that handle sessions
	SessionStore interface {
		ByID(sid string) (*Session, error)
		ByUserID(uid pgxx.ID) ([]Session, error)
		Save(s *Session) error
		Delete(s *Session) error
		DeleteByUserID(uid pgxx.ID) error
	}

	// PgxSessionStore is a redis backed session store
	RedisSessionStore struct{ redis *redis.Client }

	// PgxSessionStore is a postgres backed session store
	PgxSessionStore struct{ db pgxx.DB }

	pgxSessionRow struct {
		ID        string
		UserID    *pgxx.ID
		Data      Session
		CreatedAt time.Time
		UpdatedAt time.Time
		ExpiresAt time.Time
	}
)

func (pgxSessionRow) TableName() string {
	return "sessions"
}

// NewPgxSessionStore returns postgres backed session store
func NewPgxSessionStore(db pgxx.DB) *PgxSessionStore {
	return &PgxSessionStore{db}
}

// Init creates session table
func (s *PgxSessionStore) Init() error {
	return pgxx.Exec(s.db, `create table if not exists sessions (
		id varchar(64) primary key not null,
		user_id int,
		data jsonb not null,
		created_at timestamptz not null default now(),
		updated_at timestamptz not null default now(),
		expires_at timestamptz not null,
		foreign key (user_id) references users(id)
	)`)
}

// Drop drops the session table
func (s *PgxSessionStore) Drop() error {
	return pgxx.Exec(s.db, "drop table sessions")
}

// Save upserts session into database
func (s *PgxSessionStore) Save(sess *Session) error {
	sess.Counter++

	if sess.ID == "" {
		sid, err := GenerateToken(16)
		if err != nil {
			return err
		}

		sess.ID = sid
		return pgxx.Exec(s.db, "insert into sessions (id, user_id, data, expires_at) values ($1, $2, $3, $4)",
			sid, sess.UserID(), sess, sess.ExpiresAt)
	}

	return pgxx.Exec(s.db, "update sessions set updated_at = now(), data = $1, user_id = $2 where id = $3", sess, sess.UserID(), sess.ID)
}

// ByID returns a session by its id
func (s *PgxSessionStore) ByID(sessionID string) (*Session, error) {
	var row pgxSessionRow
	if err := pgxx.One(s.db, &row, "where id = $1", sessionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}

		return nil, err
	}

	if row.Data.Expired() {
		// we can automatically delete it
		go s.Delete(&row.Data)
		return nil, ErrSessionExpired
	}

	return &row.Data, nil
}

// ByUserID returns all sessions belonging to a user
func (s *PgxSessionStore) ByUserID(uid pgxx.ID) ([]Session, error) {
	var rows []pgxSessionRow
	if err := pgxx.Many(s.db, &rows, "user_id = $1", uid); err != nil {
		return nil, err
	}

	sessions := make([]Session, 0)
	for i := range rows {
		sess := rows[i].Data

		// delete expired sessions
		if sess.Expired() {
			go s.Delete(&sess)
			continue
		}

		sessions = append(sessions, rows[i].Data)
	}

	return sessions, nil
}

// Delete session by id
func (s *PgxSessionStore) Delete(sess *Session) error {
	return pgxx.Exec(s.db, "delete from sessions where id = $1", sess.ID)
}

// DeleteByUserID deletes all sessions for a given user
func (s *PgxSessionStore) DeleteByUserID(uid pgxx.ID) error {
	return pgxx.Exec(s.db, "delete from sessions where user_id = $1", uid)
}

// DeleteExpiredSessions deletes all sessions which have expired
func (s *PgxSessionStore) DeleteExpiredSessions() error {
	return pgxx.Exec(s.db, "delete from sessions where expires_at < now()")
}

// NewRedisSessionStore returns a redis backed session store
func NewRedisSessionStore(rd *redis.Client) *RedisSessionStore {
	return &RedisSessionStore{redis: rd}
}

// ByID returns the session with the given id from the store.
func (s RedisSessionStore) ByID(sid string) (*Session, error) {
	data, err := s.redis.Get(s.sessionKey(sid)).Result()

	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}

	if err != nil {
		return nil, err
	}

	var sess Session
	if err := json.Unmarshal([]byte(data), &sess); err != nil {
		return nil, err
	}

	if sess.Expired() {
		// remove session in the background
		go s.Delete(&sess)
		return nil, ErrSessionExpired
	}

	return &sess, nil
}

// ByUserID returns all sessions belonging to the given user
func (s RedisSessionStore) ByUserID(uid pgxx.ID) ([]Session, error) {
	key := s.userSessionKey(uid.String())
	sessionKeys, err := s.redis.SMembers(key).Result()
	if err != nil {
		return nil, err
	}

	results, err := s.redis.MGet(sessionKeys...).Result()
	if err != nil {
		return nil, err
	}

	var sessions []Session
	for i := range results {
		var sess Session
		if err := json.Unmarshal([]byte(results[i].(string)), &sess); err != nil {
			return sessions, err
		}

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

// DeleteByUserID delets all sessions for given user
func (s RedisSessionStore) DeleteByUserID(uid pgxx.ID) error {
	// we need to get all members and cascading into the otherones
	sessions, err := s.ByUserID(uid)
	if err != nil {
		return err
	}

	var keys []string
	for i := range sessions {
		keys = append(keys, sessions[i].ID)
	}

	// delete user sessions
	if err := s.redis.Del(s.userSessionKey(uid.String())).Err(); err != nil {
		return err
	}

	return s.redis.Del(keys...).Err()
}

// Delete session from cache and db
func (s RedisSessionStore) Delete(sess *Session) error {
	if err := s.redis.Del(s.sessionKey(sess.ID)).Err(); err != nil {
		return err
	}

	// cascade into user
	if !sess.IsAnonymous() {
		return s.redis.Del(s.userSessionKey(sess.UserID().String())).Err()
	}

	return nil
}

// Save session in redis store
func (s RedisSessionStore) Save(sess *Session) error {
	if sess.ID == "" {
		sid, err := GenerateToken(16)
		if err != nil {
			return err
		}

		sess.ID = sid
	}

	// increment the save counter
	sess.Counter++

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	expires := time.Until(sess.ExpiresAt)
	key := s.sessionKey(sess.ID)

	if err := s.redis.Set(key, data, expires).Err(); err != nil {
		return err
	}

	if !sess.IsAnonymous() {
		userKey := s.userSessionKey(sess.UserID().String())

		// add to set but how do we remove after?
		if err := s.redis.SAdd(userKey, key).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (RedisSessionStore) userSessionKey(uid string) string {
	return fmt.Sprintf("%s:%s", userKey, uid)
}
func (RedisSessionStore) sessionKey(id string) string {
	return fmt.Sprintf("%s:%s", sessionKey, id)
}
