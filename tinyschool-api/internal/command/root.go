package command

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"tinyschool-api/internal/config"
	"tinyschool-api/internal/server"
	"tinyschool-api/internal/service"
	"tinyschool-api/internal/storage/gormsqlite"
)

func Execute(ctx context.Context) error {
	return NewRootCommand().ExecuteContext(ctx)
}

func NewRootCommand() *cobra.Command {
	defaults := config.Default()
	values := viper.New()
	command := &cobra.Command{
		Use:           "tinyschool-api",
		Short:         "Run the Tiny School API",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg := config.Config{
				Address:         values.GetString("address"),
				DatabasePath:    values.GetString("database"),
				ShutdownTimeout: values.GetDuration("shutdown-timeout"),
				JWTSecret:       values.GetString("jwt-secret"),
				SessionDuration: values.GetDuration("session-duration"),
			}
			if err := cfg.Validate(); err != nil {
				return err
			}
			secret, err := cfg.ResolveJWTSecret()
			if err != nil {
				return err
			}

			store, err := gormsqlite.Open(cfg.DatabasePath)
			if err != nil {
				return err
			}
			defer func() {
				if err := store.Close(); err != nil {
					slog.Error("close database", "error", err)
				}
			}()
			if err := store.AutoMigrate(command.Context()); err != nil {
				return err
			}
			if err := store.Seed(command.Context()); err != nil {
				return err
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := service.New(
				store,
				service.WithJWTSecret([]byte(secret)),
				service.WithSessionDuration(cfg.SessionDuration),
			)
			return server.New(cfg.Address, cfg.ShutdownTimeout, app, logger).Run(command.Context())
		},
	}

	flags := command.Flags()
	flags.String("address", defaults.Address, "HTTP listen address")
	flags.String("database", defaults.DatabasePath, "SQLite database path")
	flags.Duration("shutdown-timeout", defaults.ShutdownTimeout, "graceful shutdown timeout")
	flags.String("jwt-secret", defaults.JWTSecret, "JWT signing secret (at least 32 bytes)")
	flags.Duration("session-duration", defaults.SessionDuration, "authenticated session duration")

	values.SetEnvPrefix("TINYSCHOOL")
	values.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	values.AutomaticEnv()
	for key, environment := range map[string]string{
		"address":          "TINYSCHOOL_API_ADDRESS",
		"database":         "TINYSCHOOL_DB_PATH",
		"jwt-secret":       "TINYSCHOOL_JWT_SECRET",
		"session-duration": "TINYSCHOOL_SESSION_DURATION",
	} {
		if err := values.BindEnv(key, environment); err != nil {
			panic(fmt.Sprintf("bind %s: %v", environment, err))
		}
	}
	if err := values.BindPFlags(flags); err != nil {
		panic(fmt.Sprintf("bind command flags: %v", err))
	}
	return command
}
