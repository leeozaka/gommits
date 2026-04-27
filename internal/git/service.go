package git

import "github.com/leeozaka/gommits/internal/models"

type GitService interface {
	IsGitRepo(path string) bool
	GetCurrentBranch(path string) (string, error)
	GetRepositoryName(path string) string
	DetectDefaultBranch(path string) string
	GatherCommits(path, author, parentBranch string, currentBranchOnly bool) ([]models.CommitInfo, string, error)
	GetChangedFiles(path, commitHash string) ([]string, error)
	PathExistsInRef(repoPath, ref, targetPath string) bool
}

type CLIGitService struct{}

func NewCLIGitService() *CLIGitService {
	return &CLIGitService{}
}

func (s *CLIGitService) IsGitRepo(path string) bool {
	return IsGitRepo(path)
}

func (s *CLIGitService) GetCurrentBranch(path string) (string, error) {
	return GetCurrentBranch(path)
}

func (s *CLIGitService) GetRepositoryName(path string) string {
	return GetRepositoryName(path)
}

func (s *CLIGitService) DetectDefaultBranch(path string) string {
	return DetectDefaultBranch(path)
}

func (s *CLIGitService) GatherCommits(path, author, parentBranch string, currentBranchOnly bool) ([]models.CommitInfo, string, error) {
	return GatherCommits(path, author, parentBranch, currentBranchOnly)
}

func (s *CLIGitService) GetChangedFiles(path, commitHash string) ([]string, error) {
	return GetChangedFiles(path, commitHash)
}

func (s *CLIGitService) PathExistsInRef(repoPath, ref, targetPath string) bool {
	return PathExistsInRef(repoPath, ref, targetPath)
}
