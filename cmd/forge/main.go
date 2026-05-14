package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/runcmd"
	"github.com/thisisthemurph/forge/internal/statuscmd"
)

func main() {
	client := &githubapi.Client{}
	deps := cli.Deps{
		Getenv: os.Getenv,
		GhAuth: ghAuthToken,
		Client: client,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getwd:  os.Getwd,
		ExecStatus: func(ctx context.Context, cfg cli.Config, dir string) error {
			return statuscmd.Run(ctx, cfg, dir, os.Getenv, ghAuthToken, client, os.Stdout)
		},
		ExecRun: func(ctx context.Context, cfg cli.Config, dir string) error {
			return runcmd.Run(ctx, cfg, dir, os.Getenv, ghAuthToken, client, os.Stdin, os.Stdout, os.Stderr)
		},
	}

	root := cli.NewRootCmd(deps)
	root.SetArgs(os.Args[1:])
	if err := root.ExecuteContext(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "forge: %v\n", err)
		os.Exit(1)
	}
}

func ghAuthToken() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
