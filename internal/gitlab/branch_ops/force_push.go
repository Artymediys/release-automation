package branch_ops

import (
	"fmt"
	"os"

	"arel/internal/gitlab/repo_ops"
	"arel/pkg/utils"

	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

func ForcePushBranch(glc *gitlab.Client, projectID, sourceBranch, targetBranch string) error {
	// Protection check
	protectedBranch, err := IsBranchProtected(glc, projectID, targetBranch)
	if err != nil {
		return err
	}

	protectedAll, err := IsBranchProtected(glc, projectID, "*")
	if err != nil {
		return err
	}

	// Allowing to Force Push
	if protectedBranch {
		err = DisableBranchProtection(glc, projectID, targetBranch)
		if err != nil {
			return err
		}
	}

	if protectedAll {
		err = DisableBranchProtection(glc, projectID, "*")
		if err != nil {
			return err
		}
	}

	// Force Push
	repoUrlProtocol, repoUrlBody, err := repo_ops.GetProjectURL(glc, projectID)
	if err != nil {
		return err
	}

	err = forcePushWithGit(repoUrlProtocol, repoUrlBody, viper.GetString("pat"), sourceBranch, targetBranch)
	if err != nil {
		return err
	}

	// Protection restoration
	if protectedBranch {
		err = RestoreBranchProtection(glc, projectID, targetBranch)
		if err != nil {
			return err
		}
	}

	if protectedAll {
		err = RestoreBranchProtection(glc, projectID, "*")
		if err != nil {
			return err
		}
	}

	return nil
}

func forcePushWithGit(repoUrlProtocol, repoUrlBody, pat, sourceBranch, targetBranch string) error {
	tmpDir, err := os.MkdirTemp("", "repo-")
	if err != nil {
		return fmt.Errorf("не удалось создать временную директорию -> %w", err)
	}
	defer os.RemoveAll(tmpDir)

	err = utils.RunGitCommand(tmpDir, "clone", fmt.Sprintf("%s://oauth2:%s@%s", repoUrlProtocol, pat, repoUrlBody), ".")
	if err != nil {
		return fmt.Errorf("не удалось клонировать репозиторий -> %w", err)
	}

	if err = utils.RunGitCommand(tmpDir, "checkout", sourceBranch); err != nil {
		return fmt.Errorf("не удалось переключиться на ветку %s -> %w", sourceBranch, err)
	}

	err = utils.RunGitCommand(tmpDir, "push", "origin", fmt.Sprintf("%s:%s", sourceBranch, targetBranch), "--force")
	if err != nil {
		return fmt.Errorf("не удалось выполнить force push с ветки %s на ветку %s -> %w", sourceBranch, targetBranch, err)
	}

	return nil
}
