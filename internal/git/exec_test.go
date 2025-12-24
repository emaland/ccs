package git

import (
	"reflect"
	"testing"
)

func TestParseWorktreeList(t *testing.T) {
	input := `worktree /Users/test/repo
HEAD abc123def456
branch refs/heads/main

worktree /Users/test/.ccs/repo/feature
HEAD def789abc012
branch refs/heads/ccs/feature

worktree /Users/test/.ccs/repo/bugfix
HEAD 123456789abc
branch refs/heads/ccs/bugfix
`
	expected := []WorktreeInfo{
		{Path: "/Users/test/repo", HEAD: "abc123def456", Branch: "main"},
		{Path: "/Users/test/.ccs/repo/feature", HEAD: "def789abc012", Branch: "ccs/feature"},
		{Path: "/Users/test/.ccs/repo/bugfix", HEAD: "123456789abc", Branch: "ccs/bugfix"},
	}

	result := parseWorktreeList(input)

	if len(result) != len(expected) {
		t.Fatalf("expected %d worktrees, got %d", len(expected), len(result))
	}

	for i, wt := range result {
		if wt.Path != expected[i].Path {
			t.Errorf("worktree %d: expected path %q, got %q", i, expected[i].Path, wt.Path)
		}
		if wt.HEAD != expected[i].HEAD {
			t.Errorf("worktree %d: expected HEAD %q, got %q", i, expected[i].HEAD, wt.HEAD)
		}
		if wt.Branch != expected[i].Branch {
			t.Errorf("worktree %d: expected branch %q, got %q", i, expected[i].Branch, wt.Branch)
		}
	}
}

func TestParseDiffStat(t *testing.T) {
	tests := []struct {
		input    string
		expected DiffStat
	}{
		{
			input:    "3 files changed, 10 insertions(+), 5 deletions(-)",
			expected: DiffStat{FilesChanged: 3, Insertions: 10, Deletions: 5},
		},
		{
			input:    "1 file changed, 1 insertion(+)",
			expected: DiffStat{FilesChanged: 1, Insertions: 1, Deletions: 0},
		},
		{
			input:    "1 file changed, 1 deletion(-)",
			expected: DiffStat{FilesChanged: 1, Insertions: 0, Deletions: 1},
		},
		{
			input:    "",
			expected: DiffStat{},
		},
		{
			input:    "10 files changed, 100 insertions(+)",
			expected: DiffStat{FilesChanged: 10, Insertions: 100, Deletions: 0},
		},
	}

	for _, tt := range tests {
		result := parseDiffStat(tt.input)
		if result.FilesChanged != tt.expected.FilesChanged {
			t.Errorf("input %q: expected FilesChanged %d, got %d", tt.input, tt.expected.FilesChanged, result.FilesChanged)
		}
		if result.Insertions != tt.expected.Insertions {
			t.Errorf("input %q: expected Insertions %d, got %d", tt.input, tt.expected.Insertions, result.Insertions)
		}
		if result.Deletions != tt.expected.Deletions {
			t.Errorf("input %q: expected Deletions %d, got %d", tt.input, tt.expected.Deletions, result.Deletions)
		}
	}
}

func TestParseDiffFiles(t *testing.T) {
	input := `M	src/main.go
A	src/new.go
D	src/old.go
R100	src/renamed.go
`
	expected := []FileChange{
		{Path: "src/main.go", Status: FileModified},
		{Path: "src/new.go", Status: FileAdded},
		{Path: "src/old.go", Status: FileDeleted},
		{Path: "src/renamed.go", Status: FileRenamed},
	}

	result := parseDiffFiles(input)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
