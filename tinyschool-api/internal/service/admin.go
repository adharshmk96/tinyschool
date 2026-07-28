package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"tinyschool-api/internal/dto"
	"tinyschool-api/internal/model"
	"tinyschool-api/internal/storage"
)

// AdminStatus tells the back office whether the first administrator still has
// to be created. It is public so the UI can route to the setup page.
func (a *App) AdminStatus(ctx context.Context) (dto.AdminStatus, error) {
	total, err := a.storage.CountAdmins(ctx)
	if err != nil {
		return dto.AdminStatus{}, err
	}
	return dto.AdminStatus{AdminExists: total > 0}, nil
}

// CreateAdmin bootstraps the first administrator. It is only reachable while no
// administrator exists, which is what keeps the public endpoint safe.
func (a *App) CreateAdmin(ctx context.Context, input dto.AdminSetupRequest) (dto.AuthResult, error) {
	status, err := a.AdminStatus(ctx)
	if err != nil {
		return dto.AuthResult{}, err
	}
	if status.AdminExists {
		return dto.AuthResult{}, conflict("an administrator already exists")
	}
	created, err := a.newAdmin(ctx, input)
	if err != nil {
		return dto.AuthResult{}, err
	}
	return a.createSession(ctx, created)
}

// AddAdmin creates a further administrator. Unlike CreateAdmin it lives behind
// the admin session guard, so only an existing administrator can call it.
func (a *App) AddAdmin(ctx context.Context, input dto.AdminSetupRequest) (dto.AdminUser, error) {
	created, err := a.newAdmin(ctx, input)
	if err != nil {
		return dto.AdminUser{}, err
	}
	return adminUserDTO(created), nil
}

func (a *App) newAdmin(ctx context.Context, input dto.AdminSetupRequest) (model.User, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Name == "" {
		input.Name = "Administrator"
	}
	if err := validEmail(input.Email); err != nil {
		return model.User{}, err
	}
	if err := validPassword(input.Password); err != nil {
		return model.User{}, err
	}
	userID, err := a.newID("adm")
	if err != nil {
		return model.User{}, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return model.User{}, err
	}
	created, err := a.storage.CreateUser(ctx, model.User{
		ID: userID, Name: input.Name, Email: input.Email,
		PasswordHash: string(hash), Role: model.RoleAdmin,
	}, nil)
	if err != nil {
		return model.User{}, translate(err, "administrator")
	}
	return created, nil
}

func (a *App) AdminLogin(ctx context.Context, input dto.LoginRequest) (dto.AuthResult, error) {
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
	if !user.IsAdmin() || user.IsBlocked() {
		return dto.AuthResult{}, unauthorized("invalid email or password")
	}
	return a.createSession(ctx, user)
}

// AuthenticateAdmin mirrors Authenticate but only accepts sessions that belong
// to an administrator, so an ordinary session cookie cannot reach /admin.
func (a *App) AuthenticateAdmin(ctx context.Context, token string) (dto.AdminUser, error) {
	sessionID, err := a.parseToken(token, false)
	if err != nil {
		return dto.AdminUser{}, err
	}
	_, user, err := a.activeSession(ctx, sessionID)
	if err != nil {
		return dto.AdminUser{}, err
	}
	if !user.IsAdmin() {
		return dto.AdminUser{}, unauthorized("administrator access required")
	}
	return adminUserDTO(user), nil
}

func (a *App) AdminRefresh(ctx context.Context, token string) (dto.AuthResult, error) {
	sessionID, err := a.parseToken(token, true)
	if err != nil {
		return dto.AuthResult{}, err
	}
	session, user, err := a.activeSession(ctx, sessionID)
	if err != nil {
		return dto.AuthResult{}, err
	}
	if !user.IsAdmin() {
		return dto.AuthResult{}, unauthorized("administrator access required")
	}
	return a.authResult(user, session)
}

func (a *App) ListUsers(ctx context.Context, input dto.ListOptions) (dto.Page[dto.AdminUser], error) {
	allowed := map[string]bool{"name": true, "email": true, "createdAt": true, "role": true}
	options, err := listOptions(input, allowed, "name")
	if err != nil {
		return dto.Page[dto.AdminUser]{}, err
	}
	items, total, err := a.storage.ListUsers(ctx, options)
	if err != nil {
		return dto.Page[dto.AdminUser]{}, err
	}
	result := make([]dto.AdminUser, len(items))
	for index := range items {
		result[index] = adminUserDTO(items[index])
	}
	return dto.Page[dto.AdminUser]{Items: result, Total: int(total), Page: options.Page, PageSize: options.PageSize}, nil
}

// SetUserBlocked blocks or unblocks an account. Blocking also revokes the
// user's live sessions so the change takes effect immediately.
func (a *App) SetUserBlocked(ctx context.Context, id string, blocked bool) (dto.AdminUser, error) {
	id = strings.TrimSpace(id)
	user, err := a.storage.UserByID(ctx, id)
	if err != nil {
		return dto.AdminUser{}, translate(err, "user")
	}
	if user.IsAdmin() {
		return dto.AdminUser{}, validation("administrator accounts cannot be blocked")
	}
	now := a.now().UTC()
	var blockedAt *time.Time
	if blocked {
		blockedAt = &now
	}
	updated, err := a.storage.SetUserBlocked(ctx, id, blockedAt)
	if err != nil {
		return dto.AdminUser{}, translate(err, "user")
	}
	if blocked {
		if err := a.storage.RevokeUserSessions(ctx, id, now); err != nil {
			return dto.AdminUser{}, err
		}
	}
	return adminUserDTO(updated), nil
}

// DeleteUser removes an account and everything it owns: schools, academic
// years, students, classes, assignments, exams and its sessions. The caller's
// own id is required so an administrator cannot delete themselves, and the last
// remaining administrator is always kept so the console stays reachable.
func (a *App) DeleteUser(ctx context.Context, actingAdminID, id string) error {
	id = strings.TrimSpace(id)
	if id == strings.TrimSpace(actingAdminID) {
		return validation("you cannot delete your own account")
	}
	user, err := a.storage.UserByID(ctx, id)
	if err != nil {
		return translate(err, "user")
	}
	if user.IsAdmin() {
		total, err := a.storage.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if total <= 1 {
			return validation("the last administrator cannot be deleted")
		}
	}
	if err := a.storage.DeleteUser(ctx, id); err != nil {
		return translate(err, "user")
	}
	return nil
}

func adminUserDTO(user model.User) dto.AdminUser {
	result := dto.AdminUser{
		ID: user.ID, Name: user.Name, Email: user.Email,
		Role: user.Role, Blocked: user.IsBlocked(),
	}
	if user.BlockedAt != nil {
		result.BlockedAt = user.BlockedAt.UTC().Format(timeFormat)
	}
	if !user.CreatedAt.IsZero() {
		result.CreatedAt = user.CreatedAt.UTC().Format(timeFormat)
	}
	return result
}
