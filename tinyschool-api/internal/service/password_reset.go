package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

// RequestPasswordReset issues a single-use reset token for the account behind
// the email address. It always reports success so the endpoint cannot be used
// to discover which addresses have accounts.
//
// SMTP is not configured yet, so the reset link is written to the application
// log instead of being emailed.
func (a *App) RequestPasswordReset(ctx context.Context, input dto.ForgotPasswordRequest) error {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if validEmail(email) != nil {
		return nil
	}
	user, err := a.storage.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil
		}
		return err
	}
	// Admins sign in from the back office and blocked accounts must stay out.
	if user.IsAdmin() || user.IsBlocked() {
		return nil
	}

	now := a.now().UTC()
	if err := a.storage.InvalidateUserPasswordResetTokens(ctx, user.ID, now); err != nil {
		return err
	}
	id, err := a.newID("prt")
	if err != nil {
		return err
	}
	secret, err := resetSecret()
	if err != nil {
		return err
	}
	_, err = a.storage.CreatePasswordResetToken(ctx, model.PasswordResetToken{
		ID:        id,
		UserID:    user.ID,
		TokenHash: hashResetToken(secret),
		ExpiresAt: now.Add(a.resetTokenDuration),
		CreatedAt: now,
	})
	if err != nil {
		return translate(err, "password reset token")
	}

	a.logger.Info("password reset requested",
		"email", user.Email,
		"userId", user.ID,
		"expiresAt", now.Add(a.resetTokenDuration).Format(timeFormat),
		"resetLink", a.resetLink(secret),
		"note", "SMTP is not configured; the link is logged instead of emailed",
	)
	return nil
}

// ResetPassword spends a reset token, replaces the password and signs every
// existing session out.
func (a *App) ResetPassword(ctx context.Context, input dto.ResetPasswordRequest) error {
	secret := strings.TrimSpace(input.Token)
	if secret == "" {
		return validation("reset token is required")
	}
	if err := validPassword(input.NewPassword); err != nil {
		return err
	}
	token, err := a.storage.PasswordResetTokenByHash(ctx, hashResetToken(secret))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return unauthorized("this reset link is invalid or has expired")
		}
		return err
	}
	now := a.now().UTC()
	if token.UsedAt != nil || !token.ExpiresAt.After(now) {
		return unauthorized("this reset link is invalid or has expired")
	}
	user, err := a.storage.UserByID(ctx, token.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return unauthorized("this reset link is invalid or has expired")
		}
		return err
	}
	if user.IsBlocked() {
		return unauthorized("this account has been blocked")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	// Spend the token first: if two requests race, only one gets past this.
	if err := a.storage.UsePasswordResetToken(ctx, token.ID, now); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return unauthorized("this reset link is invalid or has expired")
		}
		return err
	}
	user.PasswordHash = string(hash)
	if _, err := a.storage.UpdateUser(ctx, user); err != nil {
		return translate(err, "user")
	}
	return a.storage.RevokeUserSessions(ctx, user.ID, now)
}

func (a *App) resetLink(secret string) string {
	return fmt.Sprintf("%s/reset-password?token=%s", a.appBaseURL, url.QueryEscape(secret))
}

func resetSecret() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func hashResetToken(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
