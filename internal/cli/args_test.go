package cli

import (
	"testing"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		argv    []string
		want    Config
		wantErr bool
	}{
		{
			name: "status with number",
			argv: []string{"status", "11"},
			want: Config{Subcommand: "status", Feature: 11},
		},
		{
			name: "repo override",
			argv: []string{"--repo", "o/r", "status", "3"},
			want: Config{RepoOverride: "o/r", Subcommand: "status", Feature: 3},
		},
		{
			name:    "missing number",
			argv:    []string{"status"},
			wantErr: true,
		},
		{
			name:    "run missing number",
			argv:    []string{"run"},
			wantErr: true,
		},
		{
			name:    "bad number",
			argv:    []string{"status", "x"},
			wantErr: true,
		},
		{
			name: "run with number",
			argv: []string{"run", "1"},
			want: Config{Subcommand: "run", Feature: 1},
		},
		{
			name: "repo override with run",
			argv: []string{"--repo", "o/r", "run", "9"},
			want: Config{RepoOverride: "o/r", Subcommand: "run", Feature: 9},
		},
		{
			name:    "unknown command",
			argv:    []string{"deploy", "1"},
			wantErr: true,
		},
		{
			name:    "extra args",
			argv:    []string{"status", "1", "extra"},
			wantErr: true,
		},
		{
			name:    "repo without value",
			argv:    []string{"--repo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseArgs(tt.argv)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("ParseArgs(%v) = %+v, want %+v", tt.argv, got, tt.want)
			}
		})
	}
}
