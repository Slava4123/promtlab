package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	repo "promptvault/internal/interface/repository"
	"promptvault/internal/models"
)

// ===================== UpdateProfile =====================

func TestUpdateProfile_Success(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	user := &models.User{ID: 1, Name: "Old", AvatarURL: "old.png"}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	result, err := svc.UpdateProfile(ctx, 1, "New Name", "new.png", nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "new.png", result.AvatarURL)
	users.AssertExpectations(t)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	result, err := svc.UpdateProfile(ctx, 1, "Name", "avatar.png", nil)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestUpdateProfile_DBError(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	dbErr := fmt.Errorf("connection timeout")
	users.On("GetByID", ctx, uint(1)).Return(nil, dbErr)

	result, err := svc.UpdateProfile(ctx, 1, "Name", "avatar.png", nil)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
	assert.NotErrorIs(t, err, ErrUserNotFound)
}

func TestUpdateProfile_EmptyAvatarPreserved(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	user := &models.User{ID: 1, Name: "Old", AvatarURL: "existing.png"}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	result, err := svc.UpdateProfile(ctx, 1, "New Name", "", nil)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.Equal(t, "existing.png", result.AvatarURL, "empty avatarURL should NOT overwrite existing")
	users.AssertExpectations(t)
}

// ===================== ChangePassword =====================

func TestChangePassword_Success(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, PasswordHash: string(hash)}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	err := svc.ChangePassword(ctx, 1, "oldpass", "newpass")

	assert.NoError(t, err)
	// Verify new password hash is stored correctly
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newpass")))
	users.AssertExpectations(t)
}

func TestChangePassword_UserNotFound(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	err := svc.ChangePassword(ctx, 1, "old", "new")

	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestChangePassword_NoPassword(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	user := &models.User{ID: 1, PasswordHash: ""} // OAuth user — no password
	users.On("GetByID", ctx, uint(1)).Return(user, nil)

	err := svc.ChangePassword(ctx, 1, "old", "new")

	assert.ErrorIs(t, err, ErrNoPassword)
}

func TestChangePassword_WrongOldPassword(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, PasswordHash: string(hash)}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)

	err := svc.ChangePassword(ctx, 1, "wrong", "new")

	assert.ErrorIs(t, err, ErrWrongPassword)
}

// ===================== ConfirmSetPassword =====================

func TestConfirmSetPassword_Success(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, PasswordHash: ""} // OAuth user without password
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 1, UserID: 1, Code: "123456", Attempts: 0,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	verif.On("DeleteByUserID", ctx, uint(1)).Return(nil)

	err := svc.ConfirmSetPassword(ctx, 1, "123456", "newpassword")

	assert.NoError(t, err)
	assert.True(t, user.EmailVerified, "EmailVerified should be set to true")
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newpassword")))
	users.AssertExpectations(t)
}

func TestConfirmSetPassword_AlreadyHasPassword(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("existing"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, PasswordHash: string(hash)}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)

	err := svc.ConfirmSetPassword(ctx, 1, "123456", "newpassword")

	assert.ErrorIs(t, err, ErrPasswordAlreadySet)
}

func TestConfirmSetPassword_UserNotFound(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	err := svc.ConfirmSetPassword(ctx, 1, "123456", "newpassword")

	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ===================== UnlinkProvider =====================

func TestUnlinkProvider_Success(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	// 2 linked providers, no password — can unlink one
	linked.On("CountByUserID", ctx, uint(1)).Return(int64(2), nil)
	users.On("GetByID", ctx, uint(1)).Return(&models.User{ID: 1, PasswordHash: ""}, nil)
	linked.On("Delete", ctx, uint(1), "github").Return(nil)

	err := svc.UnlinkProvider(ctx, 1, "github")

	assert.NoError(t, err)
	linked.AssertExpectations(t)
}

func TestUnlinkProvider_CannotUnlinkLast_OnlyProvider(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	// Only 1 linked provider, no password — totalMethods=1, cannot unlink
	linked.On("CountByUserID", ctx, uint(1)).Return(int64(1), nil)
	users.On("GetByID", ctx, uint(1)).Return(&models.User{ID: 1, PasswordHash: ""}, nil)

	err := svc.UnlinkProvider(ctx, 1, "github")

	assert.ErrorIs(t, err, ErrCannotUnlinkLast)
}

func TestUnlinkProvider_AllowedWhenPasswordExists(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	// 1 linked provider + password — totalMethods=2, can unlink
	linked.On("CountByUserID", ctx, uint(1)).Return(int64(1), nil)
	users.On("GetByID", ctx, uint(1)).Return(&models.User{ID: 1, PasswordHash: string(hash)}, nil)
	linked.On("Delete", ctx, uint(1), "google").Return(nil)

	err := svc.UnlinkProvider(ctx, 1, "google")

	assert.NoError(t, err)
	linked.AssertExpectations(t)
}

// ===================== Login =====================

func TestLogin_Success(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, Email: "test@example.com", PasswordHash: string(hash), EmailVerified: true}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	result, tokens, err := svc.Login(ctx, "test@example.com", "password123")

	assert.NoError(t, err)
	assert.Equal(t, uint(1), result.ID)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Greater(t, tokens.ExpiresIn, int64(0))
}

func TestLogin_WrongPassword(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, Email: "test@example.com", PasswordHash: string(hash), EmailVerified: true}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

	result, tokens, err := svc.Login(ctx, "test@example.com", "wrong")

	assert.Nil(t, result)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_UserNotFound(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByEmail", ctx, "noone@example.com").Return(nil, repo.ErrNotFound)

	result, tokens, err := svc.Login(ctx, "noone@example.com", "pass")

	assert.Nil(t, result)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogin_EmailNotVerified(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, Email: "test@example.com", PasswordHash: string(hash), EmailVerified: false}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)

	result, tokens, err := svc.Login(ctx, "test@example.com", "password")

	assert.Nil(t, result)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrEmailNotVerified)
}

// ===================== Token =====================

func TestGenerateAndValidateToken(t *testing.T) {
	svc := newTestService(new(mockUserRepo), new(mockLinkedAccountRepo), new(mockVerificationRepo))

	// Generate token pair
	tokens, err := svc.generateTokenPair(42, "test-nonce")
	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)

	// Validate access token
	claims, err := svc.ValidateToken(tokens.AccessToken, TokenTypeAccess)
	assert.NoError(t, err)
	assert.Equal(t, uint(42), claims.UserID)
	assert.Equal(t, TokenTypeAccess, claims.Type)
}

func TestValidateToken_Expired(t *testing.T) {
	svc := newTestService(new(mockUserRepo), new(mockLinkedAccountRepo), new(mockVerificationRepo))

	// Generate a token that expired 1 hour ago
	now := time.Now().Add(-2 * time.Hour)
	tokenStr, err := svc.generateToken(1, TokenTypeAccess, "", now, 1*time.Hour)
	assert.NoError(t, err)

	_, err = svc.ValidateToken(tokenStr, TokenTypeAccess)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestValidateToken_WrongType(t *testing.T) {
	svc := newTestService(new(mockUserRepo), new(mockLinkedAccountRepo), new(mockVerificationRepo))

	// Generate a refresh token
	tokens, err := svc.generateTokenPair(1, "test-nonce")
	assert.NoError(t, err)

	// Try to validate refresh token as access — should fail
	_, err = svc.ValidateToken(tokens.RefreshToken, TokenTypeAccess)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

// ===================== Me =====================

func TestMe_Success(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	user := &models.User{ID: 1, Name: "John", Email: "john@example.com"}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)

	result, err := svc.Me(ctx, 1)

	assert.NoError(t, err)
	assert.Equal(t, "John", result.Name)
	assert.Equal(t, "john@example.com", result.Email)
}

func TestMe_NotFound(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	result, err := svc.Me(ctx, 1)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ===================== Register =====================

func TestRegister_NewUser(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	users.On("GetByEmail", ctx, "new@example.com").Return(nil, repo.ErrNotFound)
	users.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	result, err := svc.Register(ctx, "new@example.com", "password123", "New User", "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "new@example.com", result.Email)
	assert.Equal(t, "New User", result.Name)
	assert.False(t, result.EmailVerified)
	assert.NotEmpty(t, result.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte("password123")))
	users.AssertExpectations(t)
}

func TestRegister_ExistingVerifiedEmail(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	existing := &models.User{ID: 1, Email: "taken@example.com", EmailVerified: true}
	users.On("GetByEmail", ctx, "taken@example.com").Return(existing, nil)
	linked.On("GetByUserID", ctx, uint(1)).Return(nil, repo.ErrNotFound)

	result, err := svc.Register(ctx, "taken@example.com", "password123", "Name", "")

	assert.Nil(t, result)
	var emailErr *EmailTakenError
	assert.ErrorAs(t, err, &emailErr)
	assert.Empty(t, emailErr.Provider)
}

func TestRegister_ExistingOAuthUser(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	existing := &models.User{ID: 1, Email: "oauth@example.com", EmailVerified: true}
	users.On("GetByEmail", ctx, "oauth@example.com").Return(existing, nil)
	linked.On("GetByUserID", ctx, uint(1)).Return([]models.LinkedAccount{
		{ID: 1, UserID: 1, Provider: "github", ProviderID: "123"},
	}, nil)

	result, err := svc.Register(ctx, "oauth@example.com", "password123", "Name", "")

	assert.Nil(t, result)
	var emailErr *EmailTakenError
	assert.ErrorAs(t, err, &emailErr)
	assert.Equal(t, "GitHub", emailErr.Provider)
}

func TestRegister_ExistingUnverifiedUser(t *testing.T) {
	users := new(mockUserRepo)
	linked := new(mockLinkedAccountRepo)
	svc := newTestService(users, linked, new(mockVerificationRepo))
	ctx := context.Background()

	existing := &models.User{ID: 1, Email: "unverified@example.com", EmailVerified: false, Name: "Old Name"}
	users.On("GetByEmail", ctx, "unverified@example.com").Return(existing, nil)
	linked.On("GetByUserID", ctx, uint(1)).Return(nil, repo.ErrNotFound)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	result, err := svc.Register(ctx, "unverified@example.com", "newpass", "New Name", "")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "New Name", result.Name)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte("newpass")))
	users.AssertExpectations(t)
}

func TestRegister_DBError(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	dbErr := fmt.Errorf("db connection refused")
	users.On("GetByEmail", ctx, "new@example.com").Return(nil, repo.ErrNotFound)
	users.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(dbErr)

	result, err := svc.Register(ctx, "new@example.com", "password123", "Name", "")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, dbErr)
}

// ===================== VerifyEmail =====================

func TestVerifyEmail_Success(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com", EmailVerified: false}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "123456", Attempts: 0,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)
	verif.On("DeleteByUserID", ctx, uint(1)).Return(nil)

	resultUser, tokens, err := svc.VerifyEmail(ctx, "test@example.com", "123456")

	assert.NoError(t, err)
	assert.NotNil(t, resultUser)
	assert.True(t, resultUser.EmailVerified)
	assert.NotNil(t, tokens)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	users.AssertExpectations(t)
}

func TestVerifyEmail_ExpiredCode(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com"}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "123456", Attempts: 0,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // expired
	}, nil)

	resultUser, tokens, err := svc.VerifyEmail(ctx, "test@example.com", "123456")

	assert.Nil(t, resultUser)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrExpiredCode)
}

func TestVerifyEmail_WrongCode(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com"}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "123456", Attempts: 0,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	verif.On("IncrementAttempts", ctx, uint(10)).Return(nil)

	resultUser, tokens, err := svc.VerifyEmail(ctx, "test@example.com", "999999")

	assert.Nil(t, resultUser)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrInvalidCode)
	verif.AssertCalled(t, "IncrementAttempts", ctx, uint(10))
}

func TestVerifyEmail_TooManyAttempts(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com"}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "123456", Attempts: models.MaxVerificationAttempts,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	verif.On("DeleteByUserID", ctx, uint(1)).Return(nil)

	resultUser, tokens, err := svc.VerifyEmail(ctx, "test@example.com", "123456")

	assert.Nil(t, resultUser)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrTooManyAttempts)
	verif.AssertCalled(t, "DeleteByUserID", ctx, uint(1))
}

// ===================== Refresh =====================

func TestRefresh_Success(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	// Generate a valid refresh token first
	tokens, err := svc.generateTokenPair(1, "test-nonce")
	assert.NoError(t, err)

	user := &models.User{ID: 1, Email: "test@example.com", Name: "Test", TokenNonce: "test-nonce"}
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)

	resultUser, newTokens, err := svc.Refresh(ctx, tokens.RefreshToken)

	assert.NoError(t, err)
	assert.NotNil(t, resultUser)
	assert.Equal(t, uint(1), resultUser.ID)
	assert.NotNil(t, newTokens)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.Greater(t, newTokens.ExpiresIn, int64(0))
}

func TestRefresh_ExpiredToken(t *testing.T) {
	svc := newTestService(new(mockUserRepo), new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	// Generate a token that expired in the past
	now := time.Now().Add(-2 * time.Hour)
	expiredRefresh, err := svc.generateToken(1, TokenTypeRefresh, "test-nonce", now, 1*time.Hour)
	assert.NoError(t, err)

	resultUser, tokens, err := svc.Refresh(ctx, expiredRefresh)

	assert.Nil(t, resultUser)
	assert.Nil(t, tokens)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestRefresh_UserDeleted(t *testing.T) {
	users := new(mockUserRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), new(mockVerificationRepo))
	ctx := context.Background()

	// Generate a valid refresh token
	tokens, err := svc.generateTokenPair(999, "test-nonce")
	assert.NoError(t, err)

	users.On("GetByID", ctx, uint(999)).Return(nil, repo.ErrNotFound)

	resultUser, newTokens, err := svc.Refresh(ctx, tokens.RefreshToken)

	assert.Nil(t, resultUser)
	assert.Nil(t, newTokens)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

// ===================== ResetPassword =====================

func TestResetPassword_Success(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpassword"), bcrypt.DefaultCost)
	user := &models.User{ID: 1, Email: "test@example.com", PasswordHash: string(oldHash)}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	users.On("GetByID", ctx, uint(1)).Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "654321", Attempts: 0,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	users.On("Update", ctx, mock.AnythingOfType("*models.User")).Return(nil)
	verif.On("DeleteByUserID", ctx, uint(1)).Return(nil)

	err := svc.ResetPassword(ctx, "test@example.com", "654321", "newpassword")

	assert.NoError(t, err)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newpassword")))
	users.AssertExpectations(t)
}

func TestResetPassword_WrongCode(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com"}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "654321", Attempts: 0,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}, nil)
	verif.On("IncrementAttempts", ctx, uint(10)).Return(nil)

	err := svc.ResetPassword(ctx, "test@example.com", "000000", "newpassword")

	assert.ErrorIs(t, err, ErrInvalidCode)
	verif.AssertCalled(t, "IncrementAttempts", ctx, uint(10))
}

func TestResetPassword_ExpiredCode(t *testing.T) {
	users := new(mockUserRepo)
	verif := new(mockVerificationRepo)
	svc := newTestService(users, new(mockLinkedAccountRepo), verif)
	ctx := context.Background()

	user := &models.User{ID: 1, Email: "test@example.com"}
	users.On("GetByEmail", ctx, "test@example.com").Return(user, nil)
	verif.On("GetByUserID", ctx, uint(1)).Return(&models.EmailVerification{
		ID: 10, UserID: 1, Code: "654321", Attempts: 0,
		ExpiresAt: time.Now().Add(-1 * time.Minute), // expired
	}, nil)

	err := svc.ResetPassword(ctx, "test@example.com", "654321", "newpassword")

	assert.ErrorIs(t, err, ErrExpiredCode)
}

