package cobra_test

import (
	"bytes"
	"context"
	"fmt"
	pkgcobra "kafka-polygon/pkg/cmd/cobra"
	"kafka-polygon/pkg/converto"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
)

var emptyRunner func(_ context.Context) error

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func TestLoadFileEnv(t *testing.T) {
	t.Parallel()

	var (
		rootCmdArgs []string
		envFile     string
		expFilePath = "./test.env"
	)

	rootCmd := &cobra.Command{
		Use:               "root",
		PersistentPreRunE: pkgcobra.LoadFileEnvs(&envFile),
		Run:               func(_ *cobra.Command, args []string) { rootCmdArgs = args },
	}

	pkgcobra.LoadFlagEnv(rootCmd.PersistentFlags(), &envFile)

	output, err := executeCommand(rootCmd, "run", fmt.Sprintf("-e%s", expFilePath))
	require.NoError(t, err)

	assert.Equal(t, "", output)
	assert.Equal(t, expFilePath, envFile)

	got := strings.Join(rootCmdArgs, " ")

	assert.Equal(t, "run", got)
	assert.Equal(t, os.Getenv("SERVICE_NAME"), "test")
}

func TestLoadFileEnvs(t *testing.T) {
	t.Parallel()

	filePath := "./test.env"

	var rootCmdArgs []string

	rootCmd := &cobra.Command{
		Use:               "root",
		PersistentPreRunE: pkgcobra.LoadFileEnvs(converto.StringPointer(filePath)),
		Run:               func(_ *cobra.Command, args []string) { rootCmdArgs = args },
	}

	output, err := executeCommand(rootCmd, "one", "two")
	require.NoError(t, err)

	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")

	if got != "one two" {
		t.Errorf("rootCmdArgs expected: %q, got: %q", "one two", got)
	}

	assert.Equal(t, os.Getenv("SERVICE_NAME"), "test")
}

func TestLoadFileEnvsErr(t *testing.T) {
	t.Parallel()

	filePath := "./.env.test1"

	var rootCmdArgs []string

	rootCmd := &cobra.Command{
		Use:               "root",
		PersistentPreRunE: pkgcobra.LoadFileEnvs(converto.StringPointer(filePath)),
		Run:               func(_ *cobra.Command, args []string) { rootCmdArgs = args },
	}

	_, err := executeCommand(rootCmd, "one", "two")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "env file [./.env.test1]. open ./.env.test1: no such file or directory")
	assert.Equal(t, 0, len(rootCmdArgs))
}

func TestCommands(t *testing.T) {
	t.Parallel()

	var (
		envFile    string
		useCommads = map[string]bool{
			"consumer": true,
			"migrate":  true,
			"run":      true,
			"workflow": true,
			"cron":     true,
		}
	)

	rootCmd := pkgcobra.CmdRoot(&envFile)
	rootCmd.AddCommand(pkgcobra.CmdConsumer(emptyRunner))
	rootCmd.AddCommand(pkgcobra.CmdMigrate(emptyRunner))
	rootCmd.AddCommand(pkgcobra.CmdRunService(emptyRunner))
	rootCmd.AddCommand(pkgcobra.CmdWorkflowWorker(emptyRunner))
	rootCmd.AddCommand(pkgcobra.CmdRunCron(emptyRunner))

	for _, command := range rootCmd.Commands() {
		t.Logf("Use: %s\n", command.Use)
		check, ok := useCommads[command.Use]
		assert.True(t, ok)
		assert.True(t, check)
	}
}
