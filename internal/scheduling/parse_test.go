package scheduling

import (
	"reflect"
	"testing"
)

func TestParseBlockedBySection(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
		want []int
	}{
		{
			name: "missing section",
			body: "No blocked header here.\n",
			want: nil,
		},
		{
			name: "empty section",
			body: "## Blocked by\n\n## Next\n",
			want: nil,
		},
		{
			name: "dash bullet",
			body: "## Blocked by\n- #15\n",
			want: []int{15},
		},
		{
			name: "star bullet",
			body: "## Blocked by\n* #20\n",
			want: []int{20},
		},
		{
			name: "multiple bullets preserve order",
			body: "## Blocked by\n- #3\n- #1\n",
			want: []int{3, 1},
		},
		{
			name: "dedupe duplicate refs",
			body: "## Blocked by\n- #7\n- #7\n",
			want: []int{7},
		},
		{
			name: "section stops at next markdown heading",
			body: "## Blocked by\n- #9\n## Details\n- #99\n",
			want: []int{9},
		},
		{
			name: "multiple refs one line",
			body: "## Blocked by\n- Depends on #4 and #5\n",
			want: []int{4, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseBlockedBySection(tt.body)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseBlockedBySection() = %v, want %v", got, tt.want)
			}
		})
	}
}
