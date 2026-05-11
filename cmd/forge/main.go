package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/thisisthemurph/forge/internal/cli"
	"github.com/thisisthemurph/forge/internal/githubapi"
	"github.com/thisisthemurph/forge/internal/statuscmd"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "forge: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, argv []string, getenv func(string) string, stdout *os.File) error {
	cfg, err := cli.ParseArgs(argv)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	client := &githubapi.Client{}
	return statuscmd.Run(ctx, cfg, cwd, getenv, ghAuthToken, client, stdout)
}

func ghAuthToken() (string, error) {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
