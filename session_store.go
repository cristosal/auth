package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
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
func (s sessionStore) FindSession(id string) (*Session, error) {
	data, err := s.Get(s.keyify(id)).Result()
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
	rows, err := s.Query(ctx, "select session_id from sessions where user_id = $1", uid)
	if err != nil {
		return nil, err
	}

	sids, err := pgxx.CollectStrings(rows)
	if err != nil {
		// can be db.NotFound error
		return nil, err
	}

	sessch := make(chan Session)
	wg := new(sync.WaitGroup)
	wg.Add(len(sids))
	for i := range sids {
		go func(i int) {
			defer wg.Done()
			defer recover()
			res := s.Get(s.keyify(sids[i]))
			if res.Err() != nil {
				if errors.Is(err, redis.Nil) {
					go s.RemoveSession(sids[i])
				}
				return
			}

			data, err := res.Bytes()
			if err != nil {
				log.Printf("unable to get bytes from session: %v", err)
				return
			}

			var sess Session
			if err := json.Unmarshal(data, &sess); err != nil {
				log.Printf("unable to unmarshal session data: %v", err)
				return
			}
			sessch <- sess
		}(i)
	}

	go func() {
		wg.Wait()
		close(sessch)
	}()

	var sessions []Session
	for sess := range sessch {
		sessions = append(sessions, sess)
	}

	return sessions, nil
}

func (s sessionStore) DeleteUserSessions(uid pgxx.ID) error {
	rows, err := s.Query(ctx, "select session_id from sessions where user_id = $1", uid)
	if err != nil {
		return err
	}
	sids, err := pgxx.CollectStrings(rows)
	if err != nil {
		return err
	}

	// replace with redis session key
	for i := range sids {
		sids[i] = s.keyify(sids[i])
	}

	cmd := s.Del(sids...)
	if err := cmd.Err(); err != nil {
		return err
	}

	return pgxx.Exec(s, "delete from sessions where user_id = $1", uid)
}

func (s sessionStore) UpdateSession(sess *Session) error {
	return pgxx.Exec(s, "update sessions set user_id = $1, user_agent = $2, expires_at = $3 where session_id = $4", sess.UserID(), sess.UserAgent, sess.ExpiresAt, sess.ID)
}

// CreateSession in database and redis
func (s sessionStore) CreateSession(sess *Session) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	// persist into redis
	cmd := s.Set(s.keyify(sess.ID), data, 0)
	if err := cmd.Err(); err != nil {
		return err
	}

	return pgxx.Exec(s, "insert into sessions (session_id, user_id, user_agent, expires_at) values ($1, $2, $3, $4)", sess.ID, sess.UserID(), sess.UserAgent, sess.ExpiresAt)
}

// SaveSession session in redis
func (s sessionStore) SaveSession(sess *Session) error {
	if sess.Expired() {
		return ErrSessionExpired
	}

	// increment the save counter
	sess.Counter++

	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}

	var expires time.Duration
	if sess.Counter < 2 {
		expires = time.Minute
	} else {
		expires = time.Until(sess.ExpiresAt)
	}

	cmd := s.Set(s.keyify(sess.ID), data, expires)
	if err := cmd.Err(); err != nil {
		return err
	}

	if sess.dirty {
		sess.dirty = false
		// store user session in database
		if err := s.CreateSession(sess); err != nil {
			log.Printf("error saving session to db: %v", err)
		}
	}

	return nil
}

// RemoveSession session from cache and db
func (s sessionStore) RemoveSession(sid string) error {
	_ = s.Del(s.keyify(sid))
	return pgxx.Exec(s, "delete from sessions where session_id = $1", sid)
}

func (s sessionStore) DeleteExpiredSessions() error {
	return pgxx.Exec(s, "delete from sessions where expires_at < now()")
}

func (sessionStore) keyify(id string) string {
	return fmt.Sprintf("%s:%s", sessionKey, id)
}
