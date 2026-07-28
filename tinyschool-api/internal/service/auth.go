package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
	"tinyschool-api/internal/tenancy"
)

func (a *App) Register(ctx context.Context, input dto.RegisterRequest) (dto.AuthResult, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" {
		return dto.AuthResult{}, validation("name is required")
	}
	if err := validEmail(input.Email); err != nil {
		return dto.AuthResult{}, err
	}
	if err := validPassword(input.Password); err != nil {
		return dto.AuthResult{}, err
	}
	userID, err := a.newID("usr")
	if err != nil {
		return dto.AuthResult{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.AuthResult{}, err
	}
	ctx = tenancy.WithUserID(ctx, userID)
	created, err := a.storage.CreateUser(ctx, model.User{
		ID: userID, Name: input.Name, Email: input.Email, PasswordHash: string(hash),
	}, nil)
	if err != nil {
		return dto.AuthResult{}, translate(err, "user")
	}
	return a.createSession(ctx, created)
}

func (a *App) Login(ctx context.Context, input dto.LoginRequest) (dto.AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if err := validEmail(email); err != nil {
		return dto.AuthResult{}, unauthorized("invalid email or password")
	}
	user, err := a.storage.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return dto.AuthResult{}, unauthorized("invalid email or password")
		}
		return dto.AuthResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return dto.AuthResult{}, unauthorized("invalid email or password")
	}
	if user.IsAdmin() {
		return dto.AuthResult{}, unauthorized("administrators sign in from the admin console")
	}
	if user.IsBlocked() {
		return dto.AuthResult{}, unauthorized("this account has been blocked")
	}
	return a.createSession(ctx, user)
}

func (a *App) Refresh(ctx context.Context, token string) (dto.AuthResult, error) {
	sessionID, err := a.parseToken(token, true)
	if err != nil {
		return dto.AuthResult{}, err
	}
	session, user, err := a.activeSession(ctx, sessionID)
	if err != nil {
		return dto.AuthResult{}, err
	}
	return a.authResult(user, session)
}

func (a *App) Authenticate(ctx context.Context, token string) (dto.User, error) {
	sessionID, err := a.parseToken(token, false)
	if err != nil {
		return dto.User{}, err
	}
	_, user, err := a.activeSession(ctx, sessionID)
	if err != nil {
		return dto.User{}, err
	}
	if user.IsAdmin() {
		return dto.User{}, unauthorized("admin sessions cannot access the school workspace")
	}
	return userDTO(user), nil
}

func (a *App) Logout(ctx context.Context, token string) error {
	if strings.TrimSpace(token) == "" {
		return nil
	}
	sessionID, err := a.parseToken(token, true)
	if err != nil {
		return nil
	}
	err = a.storage.RevokeSession(ctx, sessionID, a.now().UTC())
	if errors.Is(err, storage.ErrNotFound) {
		return nil
	}
	return err
}

func (a *App) Me(ctx context.Context, userID string) (dto.User, error) {
	user, err := a.storage.UserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return dto.User{}, translate(err, "user")
	}
	return userDTO(user), nil
}

func (a *App) UpdateMe(ctx context.Context, userID string, input dto.UpdateUserRequest) (dto.User, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return dto.User{}, validation("name is required")
	}
	user, err := a.storage.UserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return dto.User{}, translate(err, "user")
	}
	user.Name = name
	updated, err := a.storage.UpdateUser(ctx, user)
	if err != nil {
		return dto.User{}, translate(err, "user")
	}
	return userDTO(updated), nil
}

func (a *App) UpdatePassword(ctx context.Context, userID, currentSessionID string, input dto.PasswordRequest) error {
	if err := validPassword(input.NewPassword); err != nil {
		return err
	}
	user, err := a.storage.UserByID(ctx, strings.TrimSpace(userID))
	if err != nil {
		return translate(err, "user")
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return unauthorized("current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hash)
	if _, err := a.storage.UpdateUser(ctx, user); err != nil {
		return translate(err, "user")
	}
	return a.storage.RevokeOtherSessions(ctx, user.ID, currentSessionID, a.now().UTC())
}

func (a *App) ClearData(ctx context.Context) error {
	if tenancy.UserID(ctx) == "" {
		return unauthorized("authentication required")
	}
	return a.storage.ClearUserData(ctx)
}

func (a *App) createSession(ctx context.Context, user model.User) (dto.AuthResult, error) {
	id, err := a.newID("ses")
	if err != nil {
		return dto.AuthResult{}, err
	}
	now := a.now().UTC()
	session, err := a.storage.CreateSession(ctx, model.Session{
		ID: id, UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(a.sessionDuration),
	})
	if err != nil {
		return dto.AuthResult{}, translate(err, "session")
	}
	return a.authResult(user, session)
}

func (a *App) activeSession(ctx context.Context, sessionID string) (model.Session, model.User, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return model.Session{}, model.User{}, unauthorized("authentication required")
	}
	session, err := a.storage.Session(ctx, sessionID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.Session{}, model.User{}, unauthorized("invalid session")
		}
		return model.Session{}, model.User{}, err
	}
	if session.RevokedAt != nil || !session.ExpiresAt.After(a.now().UTC()) {
		return model.Session{}, model.User{}, unauthorized("session expired or revoked")
	}
	user, err := a.storage.UserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return model.Session{}, model.User{}, unauthorized("invalid session")
		}
		return model.Session{}, model.User{}, err
	}
	if user.IsBlocked() {
		return model.Session{}, model.User{}, unauthorized("this account has been blocked")
	}
	return session, user, nil
}

func (a *App) authResult(user model.User, session model.Session) (dto.AuthResult, error) {
	token, _, err := a.issueToken(session)
	if err != nil {
		return dto.AuthResult{}, err
	}
	return dto.AuthResult{
		User: userDTO(user), SessionID: session.ID, ExpiresAt: session.ExpiresAt.UTC().Format(timeFormat),
		Token: token,
	}, nil
}

func userDTO(user model.User) dto.User {
	return dto.User{ID: user.ID, Name: user.Name, Email: user.Email}
}

func validEmail(value string) error {
	address, err := mail.ParseAddress(value)
	if err != nil || !strings.EqualFold(address.Address, value) {
		return validation("email must be valid")
	}
	return nil
}

func validPassword(value string) error {
	if len(value) < 8 {
		return validation("password must be at least 8 characters")
	}
	if len(value) > 72 {
		return validation("password must be at most 72 bytes")
	}
	return nil
}

const timeFormat = "2006-01-02T15:04:05Z07:00"

type tokenPayload struct {
	SessionID string `json:"sid"`
	ExpiresAt int64  `json:"exp"`
}

func (a *App) issueToken(session model.Session) (string, time.Time, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	expiresAt := a.now().UTC().Add(a.tokenDuration)
	if expiresAt.After(session.ExpiresAt) {
		expiresAt = session.ExpiresAt
	}
	payloadJSON, err := json.Marshal(tokenPayload{
		SessionID: session.ID,
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("encode token: %w", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsigned := header + "." + payload
	signature := hmac.New(sha256.New, a.jwtSecret)
	_, _ = signature.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil)), expiresAt, nil
}

func (a *App) parseToken(token string, allowExpired bool) (string, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return "", unauthorized("invalid token")
	}
	unsigned := parts[0] + "." + parts[1]
	signature := hmac.New(sha256.New, a.jwtSecret)
	_, _ = signature.Write([]byte(unsigned))
	actual, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(actual, signature.Sum(nil)) {
		return "", unauthorized("invalid token")
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", unauthorized("invalid token")
	}
	var payload tokenPayload
	if json.Unmarshal(payloadJSON, &payload) != nil || strings.TrimSpace(payload.SessionID) == "" || payload.ExpiresAt <= 0 {
		return "", unauthorized("invalid token")
	}
	if !allowExpired && a.now().UTC().Unix() >= payload.ExpiresAt {
		return "", unauthorized("token expired")
	}
	return payload.SessionID, nil
}

func (a *App) SessionID(token string, allowExpired bool) (string, error) {
	return a.parseToken(token, allowExpired)
}

func TokenExpiresAt(token string) (int64, error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token with %s parts", strconv.Itoa(len(parts)))
	}
	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, err
	}
	var payload tokenPayload
	if err := json.Unmarshal(payloadJSON, &payload); err != nil {
		return 0, err
	}
	return payload.ExpiresAt, nil
}
