package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/cristosal/pgxx"
)

const (
	sessionKey          = key("session")
	SessionDuration     = time.Hour * 3
	SessionLongDuration = time.Hour * 24 * 30
)

type (
	key string

	Session struct {
		ID          string         `json:"id"`
		Counter     int            `json:"counter"`
		User        *User          `json:"user,omitempty"`
		Permissions Permissions    `json:"permissions,omitempty"`
		Groups      Groups         `json:"groups,omitempty"`
		UserAgent   string         `json:"user_agent"`
		ExpiresAt   time.Time      `json:"expires_at"`
		Message     string         `json:"message"`
		MessageType string         `json:"message_type"`
		IP          string         `json:"ip"`
		Meta        map[string]any `json:"meta"`
		dirty       bool
	}
)

func NewSession() Session {
	sid, _ := GenerateToken(16)
	return Session{
		ID:   sid,
		Meta: make(map[string]any),
	}
}

func (s *Session) MarkDirty() {
	s.dirty = true // means save to db
}

func (s *Session) UserID() *pgxx.ID {
	if s.User == nil {
		return nil
	}

	return &s.User.ID
}

func (s *Session) Expired() bool {
	return s.ExpiresAt.Before(time.Now())
}

func (s *Session) Get(key string) any {
	return s.Meta[key]
}

func (s *Session) Set(key string, data any) {
	s.Meta[key] = data
}

func (s *Session) GroupName() string {
	if len(s.Groups) == 0 {
		return ""
	}
	return s.Groups[0].Name
}

func (s *Session) ClearFlash() {
	s.Message = ""
	s.MessageType = ""
}

func (s *Session) HasFlash() bool {
	return s.Message != ""
}

func (s *Session) Flash(msgtype, msg string) {
	s.Message = msg
	s.MessageType = msgtype
}

func (s Session) IsAnonymous() bool {
	return s.User == nil
}

func (s Session) IsAuthorized() bool {
	return !s.IsAnonymous()
}

func (s Session) MarshalBinary() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Session) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, s)
}

func GenerateToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
