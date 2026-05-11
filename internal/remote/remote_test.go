package remote

import (
	"testing"
)

func TestOwnerRepoFromURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		raw       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "https git suffix",
			raw:       "https://github.com/Acme/Widget.git",
			wantOwner: "acme",
			wantRepo:  "widget",
		},
		{
			name:      "https no suffix",
			raw:       "https://github.com/foo/bar",
			wantOwner: "foo",
			wantRepo:  "bar",
		},
		{
			name:      "scp style",
			raw:       "git@github.com:Org/Repo.git",
			wantOwner: "org",
			wantRepo:  "repo",
		},
		{
			name:      "scp no suffix",
			raw:       "git@github.com:org/repo",
			wantOwner: "org",
			wantRepo:  "repo",
		},
		{
			name:      "ssh url",
			raw:       "ssh://git@github.com/myorg/myrepo.git",
			wantOwner: "myorg",
			wantRepo:  "myrepo",
		},
		{
			name:    "not github host",
			raw:     "https://gitlab.com/foo/bar.git",
			wantErr: true,
		},
		{
			name:    "github enterprise hostname not allowed v1",
			raw:     "https://github.company.corp/foo/bar.git",
			wantErr: true,
		},
		{
			name:    "bare path",
			raw:     "foo/bar",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owner, repo, err := OwnerRepoFromURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Fatalf("OwnerRepoFromURL(%q) = (%q,%q), want (%q,%q)", tt.raw, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}

func TestParseRepoOwnerPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in        string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{"thisisthemurph/forge", "thisisthemurph", "forge", false},
		{"OWNER/REPO.git", "owner", "repo", false},
		{"bad", "", "", true},
		{"too/many/slashes", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			o, r, err := ParseRepoOwnerPath(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if o != tt.wantOwner || r != tt.wantRepo {
				t.Fatalf("ParseRepoOwnerPath(%q) = (%q,%q), want (%q,%q)", tt.in, o, r, tt.wantOwner, tt.wantRepo)
			}
		})
	}
}
