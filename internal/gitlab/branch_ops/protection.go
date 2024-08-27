package branch_ops

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

func DisableBranchProtection(glc *gitlab.Client, projectID string, branchName string) error {
	_, err := glc.ProtectedBranches.UnprotectRepositoryBranches(projectID, branchName)
	if err != nil {
		return fmt.Errorf("не удалось снять защиту с ветки %s -> %w", branchName, err)
	}

	return nil
}

func RestoreBranchProtection(client *gitlab.Client, projectID string, branchName string) error {
	allowFP := false
	noAccess := gitlab.NoPermissions
	maintainerAccess := gitlab.MaintainerPermissions
	maintainerAndDeveloperAccess := gitlab.DeveloperPermissions
	var restoreOptions *gitlab.ProtectRepositoryBranchesOptions

	switch branchName {
	case "*":
		restoreOptions = &gitlab.ProtectRepositoryBranchesOptions{
			Name:           &branchName,
			AllowForcePush: &allowFP,
			AllowedToPush:  &[]*gitlab.BranchPermissionOptions{{AccessLevel: &noAccess}},
			AllowedToMerge: &[]*gitlab.BranchPermissionOptions{{AccessLevel: &noAccess}},
		}
	case "main", "master":
		restoreOptions = &gitlab.ProtectRepositoryBranchesOptions{
			Name:           &branchName,
			AllowForcePush: &allowFP,
			AllowedToPush:  &[]*gitlab.BranchPermissionOptions{{AccessLevel: &noAccess}},
			AllowedToMerge: &[]*gitlab.BranchPermissionOptions{{AccessLevel: &maintainerAndDeveloperAccess}},
		}
	default:
		restoreOptions = &gitlab.ProtectRepositoryBranchesOptions{
			Name:           &branchName,
			AllowForcePush: &allowFP,
			AllowedToPush:  &[]*gitlab.BranchPermissionOptions{{AccessLevel: &noAccess}},
			AllowedToMerge: &[]*gitlab.BranchPermissionOptions{{AccessLevel: &maintainerAccess}},
		}
	}

	_, _, err := client.ProtectedBranches.ProtectRepositoryBranches(projectID, restoreOptions)
	if err != nil {
		return fmt.Errorf("не удалось восстановить настройки ветки -> %w", err)
	}

	return nil
}

func IsBranchProtected(glc *gitlab.Client, projectID, branchName string) (bool, error) {
	_, resp, err := glc.ProtectedBranches.GetProtectedBranch(projectID, branchName)
	if err != nil {
		if resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("не удалось проверить защиту ветки %s -> %w", branchName, err)
	}
	return true, nil
}
