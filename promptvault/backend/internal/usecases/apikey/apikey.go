package apikey

import (
	"context"
	"crypto/rand"
	"errors"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

const (
	keyPrefix      = "pvlt_"
	keyRandomBytes = 32
	maxNameLen     = 100
)

type Service struct {
	keys           repo.APIKeyRepository
	maxKeysPerUser int
}

func NewService(keys repo.APIKeyRepository, maxKeysPerUser int) *Service {
	return &Service{keys: keys, maxKeysPerUser: maxKeysPerUser}
}

func (s *Service) Create(ctx context.Context, userID uint, name string) (string, *APIKeyInfo, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil, ErrNameEmpty
	}
	if len(name) > maxNameLen {
		return "", nil, ErrNameTooLong
	}

	count, err := s.keys.CountByUserID(ctx, userID)
	if err != nil {
		return "", nil, fmt.Errorf("count api keys: %w", err)
	}
	if count >= int64(s.maxKeysPerUser) {
		return "", nil, ErrMaxKeysReached
	}

	plaintext, err := generateKey()
	if err != nil {
		return "", nil, fmt.Errorf("generate api key: %w", err)
	}

	hash := hashKey(plaintext)
	prefix := safePrefix(plaintext)

	key := &models.APIKey{
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
	}
	if err := s.keys.Create(ctx, key); err != nil {
		return "", nil, fmt.Errorf("create api key: %w", err)
	}

	info := &APIKeyInfo{
		ID:        key.ID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		CreatedAt: key.CreatedAt,
	}
	return plaintext, info, nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]APIKeyInfo, error) {
	keys, err := s.keys.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}

	result := make([]APIKeyInfo, len(keys))
	for i, k := range keys {
		result[i] = APIKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			KeyPrefix:  k.KeyPrefix,
			LastUsedAt: k.LastUsedAt,
			CreatedAt:  k.CreatedAt,
		}
	}
	return result, nil
}

func (s *Service) Revoke(ctx context.Context, keyID, userID uint) error {
	if err := s.keys.Delete(ctx, keyID, userID); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}

func (s *Service) ValidateKey(ctx context.Context, rawKey string) (*ValidateResult, error) {
	if !strings.HasPrefix(rawKey, keyPrefix) || len(rawKey) < 48 {
		return nil, ErrUnauthorized
	}

	hash := hashKey(rawKey)
	key, err := s.keys.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			slog.Warn("apikey.invalid", "prefix", safePrefix(rawKey))
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("validate api key: %w", err)
	}

	// async best-effort update last_used_at
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("apikey.update_last_used.panic", "key_id", key.ID, "recover", r)
			}
		}()
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.keys.UpdateLastUsed(bgCtx, key.ID); err != nil {
			slog.Warn("apikey.last_used_update_failed", "key_id", key.ID, "error", err)
		}
	}()

	return &ValidateResult{UserID: key.UserID, KeyID: key.ID}, nil
}

func generateKey() (string, error) {
	b := make([]byte, keyRandomBytes)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("crypto/rand unavailable: %w", err)
	}
	return keyPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// safePrefix returns the key prefix safe for logging (never more than 9 chars).
func safePrefix(key string) string {
	if len(key) >= 9 {
		return key[:9]
	}
	return key
}
