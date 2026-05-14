package cli

import (
	"fmt"
	"strconv"
)

// Config holds parsed CLI arguments for v1 forge invocations.
type Config struct {
	RepoOverride string
	Subcommand   string
	Feature      int
}

// ParseArgs parses argv without the program name (os.Args[1:]).
// Supported form: forge [--repo owner/name] (status|run) <feature-issue-number>
func ParseArgs(argv []string) (Config, error) {
	var cfg Config
	i := 0
	for i < len(argv) {
		if argv[i] != "--repo" {
			break
		}
		if i+1 >= len(argv) {
			return cfg, fmt.Errorf("--repo requires owner/name")
		}
		cfg.RepoOverride = argv[i+1]
		i += 2
	}
	if i >= len(argv) {
		return cfg, fmt.Errorf("usage: forge [--repo owner/name] (status|run) <feature-issue-number>")
	}
	switch argv[i] {
	case "status":
		cfg.Subcommand = "status"
	case "run":
		cfg.Subcommand = "run"
	default:
		return cfg, fmt.Errorf("unknown command %q (expected status or run)", argv[i])
	}
	i++
	if i >= len(argv) {
		return cfg, fmt.Errorf("%s requires feature issue number", cfg.Subcommand)
	}
	n, err := strconv.Atoi(argv[i])
	if err != nil || n < 1 {
		return cfg, fmt.Errorf("feature issue number must be a positive integer")
	}
	cfg.Feature = n
	i++
	if i < len(argv) {
		return cfg, fmt.Errorf("unexpected arguments after feature issue number")
	}
	return cfg, nil
}
