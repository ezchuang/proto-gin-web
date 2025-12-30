package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	authdomain "proto-gin-web/internal/contexts/admin/auth/domain"
)

// AdminSessionStore implements authdomain.SessionStore using Redis.
type AdminSessionStore struct {
	client    *redis.Client
	keyPrefix string
}

// NewAdminSessionStore wires a redis client into the store.
func NewAdminSessionStore(client *redis.Client) *AdminSessionStore {
	return &AdminSessionStore{
		client:    client,
		keyPrefix: "admin",
	}
}

var _ authdomain.SessionStore = (*AdminSessionStore)(nil)

func (s *AdminSessionStore) Save(ctx context.Context, session authdomain.AdminSession, ttl time.Duration) error {
	remaining := time.Until(session.AbsoluteExpiry)
	if ttl <= 0 || remaining <= 0 {
		return authdomain.ErrAdminSessionExpired
	}
	if ttl > remaining {
		ttl = remaining
	}
	payload := sessionPayload{
		ID:             session.ID,
		Profile:        session.Profile,
		IssuedAt:       session.IssuedAt,
		AbsoluteExpiry: session.AbsoluteExpiry,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	key := s.sessionKey(session.ID)
	setKey := s.userSessionsKey(session.Profile.ID)
	setTTL := remaining
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, key, data, ttl)
	pipe.SAdd(ctx, setKey, session.ID)
	pipe.Expire(ctx, setKey, setTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *AdminSessionStore) Get(ctx context.Context, id string) (authdomain.AdminSession, error) {
	key := s.sessionKey(id)
	cmd := s.client.Get(ctx, key)
	if err := cmd.Err(); err != nil {
		if err == redis.Nil {
			return authdomain.AdminSession{}, authdomain.ErrAdminSessionNotFound
		}
		return authdomain.AdminSession{}, err
	}
	var payload sessionPayload
	if err := json.Unmarshal([]byte(cmd.Val()), &payload); err != nil {
		return authdomain.AdminSession{}, err
	}
	return authdomain.AdminSession{
		ID:             payload.ID,
		Profile:        payload.Profile,
		IssuedAt:       payload.IssuedAt,
		AbsoluteExpiry: payload.AbsoluteExpiry,
	}, nil
}

func (s *AdminSessionStore) Touch(ctx context.Context, session authdomain.AdminSession, ttl time.Duration) error {
	remaining := time.Until(session.AbsoluteExpiry)
	if ttl > remaining {
		ttl = remaining
	}
	if ttl <= 0 {
		return s.Delete(ctx, session.ID)
	}
	key := s.sessionKey(session.ID)
	setKey := s.userSessionsKey(session.Profile.ID)
	setTTL := ttl
	pipe := s.client.TxPipeline()
	pipe.Expire(ctx, key, ttl)
	pipe.Expire(ctx, setKey, setTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *AdminSessionStore) Delete(ctx context.Context, id string) error {
	session, err := s.Get(ctx, id)
	if err != nil && !errors.Is(err, authdomain.ErrAdminSessionNotFound) {
		return err
	}
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.sessionKey(id))
	if err == nil {
		pipe.SRem(ctx, s.userSessionsKey(session.Profile.ID), id)
	}
	_, execErr := pipe.Exec(ctx)
	return execErr
}

func (s *AdminSessionStore) DeleteByUser(ctx context.Context, userID int64) error {
	setKey := s.userSessionsKey(userID)
	sessionIDs, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	pipe := s.client.TxPipeline()
	for _, id := range sessionIDs {
		pipe.Del(ctx, s.sessionKey(id))
	}
	pipe.Del(ctx, setKey)
	_, execErr := pipe.Exec(ctx)
	return execErr
}

type sessionPayload struct {
	ID             string           `json:"id"`
	Profile        authdomain.Admin `json:"profile"`
	IssuedAt       time.Time        `json:"issued_at"`
	AbsoluteExpiry time.Time        `json:"absolute_expiry"`
}

func (s *AdminSessionStore) sessionKey(id string) string {
	return fmt.Sprintf("%s:sessions:%s", s.keyPrefix, id)
}

func (s *AdminSessionStore) userSessionsKey(userID int64) string {
	return fmt.Sprintf("%s:user:%d:sessions", s.keyPrefix, userID)
}

