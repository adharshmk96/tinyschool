package service

import (
	"context"
	"errors"
	"fmt"

	"tinyschool-api/internal/backup"
)

func (a *App) BackupSettings(ctx context.Context) (backup.Settings, error) {
	if a.backups == nil {
		return backup.Settings{}, fmt.Errorf("backup manager is not configured")
	}
	return a.backups.Settings(ctx)
}

func (a *App) SaveBackupSettings(ctx context.Context, input backup.Settings) (backup.Settings, error) {
	if a.backups == nil {
		return backup.Settings{}, fmt.Errorf("backup manager is not configured")
	}
	value, err := a.backups.SaveSettings(ctx, input)
	if err != nil {
		return backup.Settings{}, validation(err.Error())
	}
	return value, nil
}

func (a *App) ListBackups() ([]backup.File, error) {
	if a.backups == nil {
		return nil, fmt.Errorf("backup manager is not configured")
	}
	return a.backups.List()
}

func (a *App) CreateBackup(ctx context.Context) (backup.File, error) {
	if a.backups == nil {
		return backup.File{}, fmt.Errorf("backup manager is not configured")
	}
	value, err := a.backups.Create(ctx)
	return value, translateBackupError(err)
}

func (a *App) BackupDownload(name string) (string, backup.File, error) {
	if a.backups == nil {
		return "", backup.File{}, fmt.Errorf("backup manager is not configured")
	}
	path, value, err := a.backups.Path(name)
	return path, value, translateBackupError(err)
}

func (a *App) RestoreBackup(ctx context.Context, name string) (backup.File, error) {
	if a.backups == nil {
		return backup.File{}, fmt.Errorf("backup manager is not configured")
	}
	value, err := a.backups.Restore(ctx, name)
	return value, translateBackupError(err)
}

func translateBackupError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, backup.ErrBusy):
		return conflict("another backup operation is already running")
	case errors.Is(err, backup.ErrNotFound):
		return &Error{Kind: ErrNotFound, Message: "backup not found", Err: err}
	default:
		return err
	}
}
