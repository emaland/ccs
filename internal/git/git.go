package git

// FileStatus represents the status of a file in git
type FileStatus string

const (
	FileAdded    FileStatus = "A"
	FileModified FileStatus = "M"
	FileDeleted  FileStatus = "D"
	FileRenamed  FileStatus = "R"
	FileCopied   FileStatus = "C"
)

// WorktreeInfo contains information about a git worktree
type WorktreeInfo struct {
	Path   string
	Branch string
	HEAD   string
	Bare   bool
}

// DiffStat contains diff statistics
type DiffStat struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

// FileChange represents a changed file
type FileChange struct {
	Path   string
	Status FileStatus
}

// Git is the interface for git operations
type Git interface {
	// RepoRoot returns the repository root path
	RepoRoot() string

	// RepoName returns the repository name (directory name)
	RepoName() string

	// Worktree operations
	WorktreeAdd(path, branch, base string) error
	WorktreeList() ([]WorktreeInfo, error)
	WorktreeRemove(path string, force bool) error

	// Branch operations
	BranchCreate(name, ref string) error
	BranchDelete(name string, force bool) error
	BranchCurrent() (string, error)
	BranchExists(name string) bool

	// Ref operations
	ResolveRef(ref string) (string, error)
	MergeBase(ref1, ref2 string) (string, error)

	// Diff and status
	DiffStat(base, head string) (*DiffStat, error)
	DiffFiles(base, head string) ([]FileChange, error)
	DiffRaw(base, head string, args ...string) (string, error)
	CommitCount(base, head string) (int, error)
	IsClean() (bool, error)

	// Log
	Log(base, head string, args ...string) (string, error)

	// Commit operations
	MergeSquash(branch string) error
	Merge(branch string, ff bool) error
	Commit(message string) error

	// Remote
	Push(branch string, force bool) error
	RemoteURL(name string) (string, error)

	// Working in different directories
	InWorktree(path string) Git
}
