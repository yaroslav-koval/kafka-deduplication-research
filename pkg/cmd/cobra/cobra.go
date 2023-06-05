package cobra

import (
	"context"
	"kafka-polygon/pkg/cerror"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Runner func(ctx context.Context) error

func CmdRoot(envFile *string) *cobra.Command {
	return &cobra.Command{PersistentPreRunE: LoadFileEnvs(envFile)}
}

func CmdMigrate(r Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "run migration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return r(cmd.Context())
		},
	}
}

func CmdRunService(r Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "run service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return r(cmd.Context())
		},
	}
}

func CmdRunCron(r Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "cron",
		Short: "run service related cron jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return r(cmd.Context())
		},
	}
}

func CmdWorkflowWorker(r Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "workflow",
		Short: "run workflow worker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return r(cmd.Context())
		},
	}
}

func CmdConsumer(r Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "consumer",
		Short: "run consumer",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return r(cmd.Context())
		},
	}
}

func LoadFlagEnv(fs *pflag.FlagSet, dst *string) {
	fs.StringVarP(dst, "env", "e", "", "Path to env file")
}

func LoadFileEnvs(envFile *string) func(_ *cobra.Command, _ []string) error {
	return func(_ *cobra.Command, _ []string) error {
		if envFile != nil && *envFile != "" {
			if err := godotenv.Load(*envFile); err != nil {
				return cerror.NewF(
					context.Background(),
					cerror.KindInternal,
					"load env file [%s]. %s", *envFile, err.Error(),
				).LogError()
			}
		}

		return nil
	}
}
