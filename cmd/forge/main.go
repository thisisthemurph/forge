package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/thisisthemurph/forge/internal/githubapi"
)

func main() {
	client := &githubapi.Client{}
	deps := &forgeDeps{
		Getenv: os.Getenv,
		GhAuth: ghAuthToken,
		Client: client,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Getwd:  os.Getwd,
	}

	root := newRootCmd(deps)
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
