package command

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"tinyschool-api/internal/storage/gormsqlite"
)

// newSeedCommand fills an existing account with demo data. Seeding is never
// automatic; it only happens when this command is run for a specific email.
func newSeedCommand(values *viper.Viper) *cobra.Command {
	command := &cobra.Command{
		Use:           "seed <email>",
		Short:         "Seed demo students, classes, assignments and exams for an account",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(command *cobra.Command, args []string) error {
			path := values.GetString("database")
			if database, _ := command.Flags().GetString("database"); database != "" {
				path = database
			}
			store, err := gormsqlite.Open(path)
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
			if err := store.SeedUserData(command.Context(), args[0]); err != nil {
				return err
			}
			fmt.Fprintf(command.OutOrStdout(), "seeded demo data for %s\n", args[0])
			return nil
		},
	}
	command.Flags().String("database", "", "SQLite database path (defaults to the server setting)")
	return command
}
