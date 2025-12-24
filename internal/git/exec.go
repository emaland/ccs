package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ExecGit implements Git interface using the git CLI
type ExecGit struct {
	repoRoot string
}

// NewExecGit creates a new ExecGit instance
func NewExecGit(repoRoot string) (*ExecGit, error) {
	absPath, err := filepath.Abs(repoRoot)
	if err != nil {
		return nil, err
	}
	return &ExecGit{repoRoot: absPath}, nil
}

// FindRepoRoot finds the git repository root from the given path
func FindRepoRoot(path string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

func (g *ExecGit) git(args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.repoRoot
	return cmd
}

func (g *ExecGit) gitOutput(args ...string) (string, error) {
	cmd := g.git(args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (g *ExecGit) RepoRoot() string {
	return g.repoRoot
}

func (g *ExecGit) RepoName() string {
	return filepath.Base(g.repoRoot)
}

func (g *ExecGit) WorktreeAdd(path, branch, base string) error {
	_, err := g.gitOutput("worktree", "add", "-b", branch, path, base)
	return err
}

func (g *ExecGit) WorktreeList() ([]WorktreeInfo, error) {
	out, err := g.gitOutput("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseWorktreeList(out), nil
}

func parseWorktreeList(output string) []WorktreeInfo {
	var worktrees []WorktreeInfo
	var current *WorktreeInfo

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				worktrees = append(worktrees, *current)
			}
			current = &WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "HEAD ") && current != nil {
			current.HEAD = strings.TrimPrefix(line, "HEAD ")
		} else if strings.HasPrefix(line, "branch ") && current != nil {
			// Branch is refs/heads/name, extract just the name
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			current.Branch = branch
		} else if line == "bare" && current != nil {
			current.Bare = true
		}
	}

	if current != nil {
		worktrees = append(worktrees, *current)
	}

	return worktrees
}

func (g *ExecGit) WorktreeRemove(path string, force bool) error {
	args := []string{"worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)
	_, err := g.gitOutput(args...)
	return err
}

func (g *ExecGit) BranchCreate(name, ref string) error {
	_, err := g.gitOutput("branch", name, ref)
	return err
}

func (g *ExecGit) BranchDelete(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := g.gitOutput("branch", flag, name)
	return err
}

func (g *ExecGit) BranchCurrent() (string, error) {
	return g.gitOutput("rev-parse", "--abbrev-ref", "HEAD")
}

func (g *ExecGit) BranchExists(name string) bool {
	_, err := g.gitOutput("rev-parse", "--verify", "refs/heads/"+name)
	return err == nil
}

func (g *ExecGit) ResolveRef(ref string) (string, error) {
	return g.gitOutput("rev-parse", ref)
}

func (g *ExecGit) MergeBase(ref1, ref2 string) (string, error) {
	return g.gitOutput("merge-base", ref1, ref2)
}

func (g *ExecGit) DiffStat(base, head string) (*DiffStat, error) {
	out, err := g.gitOutput("diff", "--shortstat", base+".."+head)
	if err != nil {
		return nil, err
	}
	return parseDiffStat(out), nil
}

func parseDiffStat(output string) *DiffStat {
	stat := &DiffStat{}
	if output == "" {
		return stat
	}

	// Pattern: "3 files changed, 10 insertions(+), 5 deletions(-)"
	filesRe := regexp.MustCompile(`(\d+) files? changed`)
	insertRe := regexp.MustCompile(`(\d+) insertions?\(\+\)`)
	deleteRe := regexp.MustCompile(`(\d+) deletions?\(-\)`)

	if m := filesRe.FindStringSubmatch(output); m != nil {
		stat.FilesChanged, _ = strconv.Atoi(m[1])
	}
	if m := insertRe.FindStringSubmatch(output); m != nil {
		stat.Insertions, _ = strconv.Atoi(m[1])
	}
	if m := deleteRe.FindStringSubmatch(output); m != nil {
		stat.Deletions, _ = strconv.Atoi(m[1])
	}

	return stat
}

func (g *ExecGit) DiffFiles(base, head string) ([]FileChange, error) {
	out, err := g.gitOutput("diff", "--name-status", base+".."+head)
	if err != nil {
		return nil, err
	}
	return parseDiffFiles(out), nil
}

func parseDiffFiles(output string) []FileChange {
	var changes []FileChange
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			status := FileStatus(parts[0][:1]) // Take first char (R100 -> R)
			path := parts[len(parts)-1]        // Use last part (for renames, this is the new name)
			changes = append(changes, FileChange{Path: path, Status: status})
		}
	}
	return changes
}

func (g *ExecGit) DiffRaw(base, head string, args ...string) (string, error) {
	cmdArgs := []string{"diff"}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, base+".."+head)
	return g.gitOutput(cmdArgs...)
}

func (g *ExecGit) CommitCount(base, head string) (int, error) {
	out, err := g.gitOutput("rev-list", "--count", base+".."+head)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(out)
}

func (g *ExecGit) IsClean() (bool, error) {
	out, err := g.gitOutput("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

func (g *ExecGit) Log(base, head string, args ...string) (string, error) {
	cmdArgs := []string{"log"}
	cmdArgs = append(cmdArgs, args...)
	cmdArgs = append(cmdArgs, base+".."+head)
	return g.gitOutput(cmdArgs...)
}

func (g *ExecGit) MergeSquash(branch string) error {
	_, err := g.gitOutput("merge", "--squash", branch)
	return err
}

func (g *ExecGit) Merge(branch string, ff bool) error {
	args := []string{"merge"}
	if !ff {
		args = append(args, "--no-ff")
	}
	args = append(args, branch)
	_, err := g.gitOutput(args...)
	return err
}

func (g *ExecGit) Commit(message string) error {
	_, err := g.gitOutput("commit", "-m", message)
	return err
}

func (g *ExecGit) Push(branch string, force bool) error {
	args := []string{"push", "-u", "origin"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, branch)
	_, err := g.gitOutput(args...)
	return err
}

func (g *ExecGit) RemoteURL(name string) (string, error) {
	return g.gitOutput("remote", "get-url", name)
}

func (g *ExecGit) InWorktree(path string) Git {
	return &ExecGit{repoRoot: path}
}
