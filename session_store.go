package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cristosal/pgxx"
	"github.com/go-redis/redis/v7"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type (
	SessionStore interface {
		FindSession(id string) (*Session, error)
		CreateSession(sess *Session) error
		SaveSession(sess *Session) error
		UserSessions(uid pgxx.ID) ([]Session, error)
		DeleteUserSessions(uid pgxx.ID) error
		RemoveSession(id string) error
	}

	sessionStore struct {
		*redis.Client
		pgxx.DB
	}
)

// Creates a new redis Store
func NewSessionStore(conn pgxx.DB, rd *redis.Client) SessionStore {
	return &sessionStore{
		rd,
		conn,
	}
}

// FindSession returns the session with the given id from the store.
func (s sessionStore) FindSession(sid string) (*Session, error) {
	data, err := s.Get(s.sessionKey(sid)).Result()

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
		go s.RemoveSession(sess.ID)
		return nil, ErrSessionExpired
	}

	return &sess, nil
}

func (s sessionStore) UserSessions(uid pgxx.ID) ([]Session, error) {
	key := s.userSessionKey(uid.String())
	sessionKeys, err := s.SMembers(key).Result()
	if err != nil {
		return nil, err
	}

	results, err := s.MGet(sessionKeys...).Result()
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

func (s sessionStore) DeleteUserSessions(uid pgxx.ID) error {
	return s.Del(s.userSessionKey(uid.String())).Err()
}

func (s sessionStore) UpdateSession(sess *Session) error {
	return pgxx.Exec(s, "update sessions set user_id = $1, user_agent = $2, expires_at = $3 where session_id = $4", sess.UserID(), sess.UserAgent, sess.ExpiresAt, sess.ID)
}

// CreateSession adds session to database and redis
func (s sessionStore) CreateSession(sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	// persist into redis
	cmd := s.Set(s.sessionKey(sess.ID), data, 0)
	if err := cmd.Err(); err != nil {
		return err
	}

	return pgxx.Exec(s, "insert into sessions (session_id, user_id, user_agent, expires_at) values ($1, $2, $3, $4)", sess.ID, sess.UserID(), sess.UserAgent, sess.ExpiresAt)
}

// SaveSession session in redis
func (s sessionStore) SaveSession(sess *Session) error {
	if sess.ID == "" {
		sid, err := GenerateToken(16)
		if err != nil {
			return err
		}

		sess.ID = sid
	}

	if sess.Expired() {
		return ErrSessionExpired
	}

	// increment the save counter
	sess.Counter++

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	expires := time.Until(sess.ExpiresAt)
	key := s.sessionKey(sess.ID)

	if err := s.Set(key, data, expires).Err(); err != nil {
		return err
	}

	if !sess.IsAnonymous() {
		userKey := s.userSessionKey(sess.UserID().String())

		// add to set but how do we remove after?
		if err := s.SAdd(userKey, key).Err(); err != nil {
			return err
		}
	}

	return nil
}

// RemoveSession session from cache and db
func (s sessionStore) RemoveSession(sid string) error {
	_ = s.Del(s.sessionKey(sid))
	return pgxx.Exec(s, "delete from sessions where session_id = $1", sid)
}

func (s sessionStore) DeleteExpiredSessions() error {
	return pgxx.Exec(s, "delete from sessions where expires_at < now()")
}

func (sessionStore) userSessionKey(uid string) string {
	return fmt.Sprintf("%s:%s", userKey, uid)
}

func (sessionStore) sessionKey(id string) string {
	return fmt.Sprintf("%s:%s", sessionKey, id)
}
