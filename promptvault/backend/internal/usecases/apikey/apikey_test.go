package apikey

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// --- mock ---

type mockAPIKeyRepo struct{ mock.Mock }

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *models.APIKey) error {
	return m.Called(ctx, key).Error(0)
}
func (m *mockAPIKeyRepo) ListByUserID(ctx context.Context, userID uint) ([]models.APIKey, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.APIKey), args.Error(1)
}
func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*models.APIKey, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}
func (m *mockAPIKeyRepo) Delete(ctx context.Context, id, userID uint) error {
	return m.Called(ctx, id, userID).Error(0)
}
func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockAPIKeyRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// --- helpers ---

func newTestService(keys *mockAPIKeyRepo) *Service {
	return NewService(keys, 5)
}

// --- Create ---

func TestCreate_Success(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("CountByUserID", ctx, uint(1)).Return(int64(0), nil)
	keys.On("Create", ctx, mock.AnythingOfType("*models.APIKey")).Return(nil)

	plaintext, info, err := svc.Create(ctx, CreateInput{UserID: 1, Name: "Test Key"})

	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(plaintext, "pvlt_"))
	assert.True(t, len(plaintext) >= 48)
	assert.Equal(t, "Test Key", info.Name)
	assert.True(t, strings.HasPrefix(info.KeyPrefix, "pvlt_"))
	keys.AssertExpectations(t)
}

func TestCreate_EmptyName(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)

	_, _, err := svc.Create(context.Background(), CreateInput{UserID: 1, Name: ""})
	assert.ErrorIs(t, err, ErrNameEmpty)

	_, _, err = svc.Create(context.Background(), CreateInput{UserID: 1, Name: "   "})
	assert.ErrorIs(t, err, ErrNameEmpty)
}

func TestCreate_NameTooLong(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)

	longName := strings.Repeat("a", 101)
	_, _, err := svc.Create(context.Background(), CreateInput{UserID: 1, Name: longName})
	assert.ErrorIs(t, err, ErrNameTooLong)
}

func TestCreate_MaxKeysReached(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("CountByUserID", ctx, uint(1)).Return(int64(5), nil)

	_, _, err := svc.Create(ctx, CreateInput{UserID: 1, Name: "Another Key"})
	assert.ErrorIs(t, err, ErrMaxKeysReached)
	keys.AssertNotCalled(t, "Create")
}

func TestCreate_TwoCallsProduceDifferentKeys(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("CountByUserID", ctx, uint(1)).Return(int64(0), nil)
	keys.On("Create", ctx, mock.AnythingOfType("*models.APIKey")).Return(nil)

	key1, _, err1 := svc.Create(ctx, CreateInput{UserID: 1, Name: "Key 1"})
	key2, _, err2 := svc.Create(ctx, CreateInput{UserID: 1, Name: "Key 2"})

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, key1, key2)
}

// --- ValidateKey ---

func TestValidateKey_Success(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	plaintext, _ := generateKey()
	hash := hashKey(plaintext)

	keys.On("GetByHash", ctx, hash).Return(&models.APIKey{
		ID:     1,
		UserID: 42,
	}, nil)
	keys.On("UpdateLastUsed", mock.Anything, uint(1)).Return(nil)

	result, err := svc.ValidateKey(ctx, plaintext)

	assert.NoError(t, err)
	assert.Equal(t, uint(42), result.UserID)
	assert.Equal(t, uint(1), result.KeyID)
}

func TestValidateKey_MalformedPrefix(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)

	_, err := svc.ValidateKey(context.Background(), "invalid_key_no_prefix")
	assert.ErrorIs(t, err, ErrUnauthorized)
	keys.AssertNotCalled(t, "GetByHash")
}

func TestValidateKey_TooShort(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)

	_, err := svc.ValidateKey(context.Background(), "pvlt_short")
	assert.ErrorIs(t, err, ErrUnauthorized)
	keys.AssertNotCalled(t, "GetByHash")
}

func TestValidateKey_WrongHash(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	// valid format but not in DB
	fakeKey := "pvlt_" + strings.Repeat("a", 43)
	hash := hashKey(fakeKey)

	keys.On("GetByHash", ctx, hash).Return(nil, repo.ErrNotFound)

	_, err := svc.ValidateKey(ctx, fakeKey)
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func TestValidateKey_UpdateLastUsedError_Ignored(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	plaintext, _ := generateKey()
	hash := hashKey(plaintext)

	keys.On("GetByHash", ctx, hash).Return(&models.APIKey{
		ID:     1,
		UserID: 42,
	}, nil)
	keys.On("UpdateLastUsed", mock.Anything, uint(1)).Return(errors.New("db error"))

	result, err := svc.ValidateKey(ctx, plaintext)

	assert.NoError(t, err)
	assert.Equal(t, uint(42), result.UserID)
}

// --- ValidateKey expiry (HIGH-5) ---

func TestValidateKey_ExpiredKey(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	plaintext, _ := generateKey()
	hash := hashKey(plaintext)
	expired := time.Now().Add(-time.Hour)

	keys.On("GetByHash", ctx, hash).Return(&models.APIKey{
		ID: 1, UserID: 42, ExpiresAt: &expired,
	}, nil)

	_, err := svc.ValidateKey(ctx, plaintext)
	assert.ErrorIs(t, err, ErrExpired)
	// UpdateLastUsed не должен вызываться при истёкшем ключе
	keys.AssertNotCalled(t, "UpdateLastUsed")
}

func TestValidateKey_FutureExpiry(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	plaintext, _ := generateKey()
	hash := hashKey(plaintext)
	future := time.Now().Add(time.Hour)

	keys.On("GetByHash", ctx, hash).Return(&models.APIKey{
		ID: 1, UserID: 42, ExpiresAt: &future,
	}, nil)
	keys.On("UpdateLastUsed", mock.Anything, uint(1)).Return(nil)

	result, err := svc.ValidateKey(ctx, plaintext)
	assert.NoError(t, err)
	assert.Equal(t, uint(42), result.UserID)
	assert.NotNil(t, result.Policy.ExpiresAt)
}

func TestValidateKey_NilExpiryWorks(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	plaintext, _ := generateKey()
	hash := hashKey(plaintext)

	keys.On("GetByHash", ctx, hash).Return(&models.APIKey{
		ID: 1, UserID: 42, ExpiresAt: nil,
	}, nil)
	keys.On("UpdateLastUsed", mock.Anything, uint(1)).Return(nil)

	_, err := svc.ValidateKey(ctx, plaintext)
	assert.NoError(t, err)
}

// --- Create validations (HIGH-5) ---

func TestCreate_ExpiresInPast(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	past := time.Now().Add(-time.Hour)

	_, _, err := svc.Create(context.Background(), CreateInput{
		UserID:    1,
		Name:      "Test",
		ExpiresAt: &past,
	})
	assert.ErrorIs(t, err, ErrInvalidExpires)
	keys.AssertNotCalled(t, "Create")
}

func TestCreate_UnknownTool(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)

	_, _, err := svc.Create(context.Background(), CreateInput{
		UserID:       1,
		Name:         "Test",
		AllowedTools: []string{"unknown_tool"},
	})
	assert.ErrorIs(t, err, ErrInvalidToolName)
	keys.AssertNotCalled(t, "Create")
}

// TestIsKnownTool_V12Tools защищает от регрессии, когда в mcpserver/tools.go
// регистрируется новый tool, но его забывают добавить в KnownTools. Без этого
// создание scoped API-key с новым tool'ом падает с ErrInvalidToolName.
func TestIsKnownTool_V12Tools(t *testing.T) {
	v12Tools := []string{"list_teams", "whoami", "list_trash", "restore_prompt", "purge_prompt"}
	for _, name := range v12Tools {
		assert.Truef(t, IsKnownTool(name), "tool %q must be in KnownTools whitelist", name)
	}
}

func TestCreate_ValidScope(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("CountByUserID", ctx, uint(1)).Return(int64(0), nil)
	keys.On("Create", ctx, mock.AnythingOfType("*models.APIKey")).Return(nil)

	teamID := uint(42)
	future := time.Now().Add(24 * time.Hour)
	_, info, err := svc.Create(ctx, CreateInput{
		UserID:       1,
		Name:         "Scoped",
		ReadOnly:     true,
		TeamID:       &teamID,
		AllowedTools: []string{"list_prompts", "get_prompt"},
		ExpiresAt:    &future,
	})
	assert.NoError(t, err)
	assert.True(t, info.ReadOnly)
	assert.Equal(t, &teamID, info.TeamID)
	assert.Len(t, info.AllowedTools, 2)
	assert.NotNil(t, info.ExpiresAt)
}

// --- Revoke ---

func TestRevoke_Success(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("Delete", ctx, uint(1), uint(42)).Return(nil)

	err := svc.Revoke(ctx, 1, 42)
	assert.NoError(t, err)
	keys.AssertExpectations(t)
}

func TestRevoke_NotFound(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("Delete", ctx, uint(1), uint(42)).Return(repo.ErrNotFound)

	err := svc.Revoke(ctx, 1, 42)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

func TestRevoke_WrongUser(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	// repo returns ErrNotFound because WHERE user_id doesn't match
	keys.On("Delete", ctx, uint(1), uint(99)).Return(repo.ErrNotFound)

	err := svc.Revoke(ctx, 1, 99)
	assert.ErrorIs(t, err, ErrKeyNotFound)
}

// --- List ---

func TestList_Success(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("ListByUserID", ctx, uint(1)).Return([]models.APIKey{
		{ID: 1, Name: "Key 1", KeyPrefix: "pvlt_aB3x"},
		{ID: 2, Name: "Key 2", KeyPrefix: "pvlt_cD4y"},
	}, nil)

	result, err := svc.List(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Key 1", result[0].Name)
}

func TestList_Empty(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("ListByUserID", ctx, uint(1)).Return([]models.APIKey{}, nil)

	result, err := svc.List(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

func TestList_DBError(t *testing.T) {
	keys := new(mockAPIKeyRepo)
	svc := newTestService(keys)
	ctx := context.Background()

	keys.On("ListByUserID", ctx, uint(1)).Return([]models.APIKey(nil), errors.New("db error"))

	_, err := svc.List(ctx, 1)
	assert.Error(t, err)
}

// --- safePrefix ---

func TestSafePrefix(t *testing.T) {
	assert.Equal(t, "pvlt_aB3x", safePrefix("pvlt_aB3xK9mNlongkey"))
	assert.Equal(t, "short", safePrefix("short"))
	assert.Equal(t, "", safePrefix(""))
}
