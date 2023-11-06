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
		ByID(sid string) (*Session, error)
		ByUserID(uid pgxx.ID) ([]Session, error)
		Save(s *Session) error
		Delete(s *Session) error
		DeleteByUserID(uid pgxx.ID) error
	}

	RedisSessionStore struct {
		redis *redis.Client
	}
)

// Creates a new redis backed session store
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
