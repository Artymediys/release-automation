package version_ops

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"time"

	"arel/internal/gitlab/branch_ops"

	"github.com/xanzy/go-gitlab"
)

func CheckAndUpdateVersion(
	glc *gitlab.Client,
	projectID, branch, comment, updateBuildVersion string,
	updateFullVersion *string,
) error {
	content, err := getChangelogContent(glc, projectID, branch)
	if err != nil {
		return err
	}

	lastVersion := getLastVersion(content)
	newVersion, err := generateNewVersion(lastVersion, *updateFullVersion, updateBuildVersion)
	if err != nil {
		return err
	}
	*updateFullVersion = newVersion

	if comment == "" {
		osUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("не удалось получить текущего пользователя ОС -> %w", err)
		}
		comment = fmt.Sprintf(
			"Новая версия %s от %s, запушено юзером %s",
			newVersion,
			time.Now().Format("2006-01-02"),
			osUser.Username,
		)
	}

	protectedBranch, err := branch_ops.IsBranchProtected(glc, projectID, branch)
	if err != nil {
		return err
	}

	protectedAll, err := branch_ops.IsBranchProtected(glc, projectID, "*")
	if err != nil {
		return err
	}

	if protectedBranch {
		err = branch_ops.DisableBranchProtection(glc, projectID, branch)
		if err != nil {
			return err
		}
	}

	if protectedAll {
		err = branch_ops.DisableBranchProtection(glc, projectID, "*")
		if err != nil {
			return err
		}
	}

	newContent := insertNewVersion(content, newVersion, comment)
	err = updateChangelogContent(glc, projectID, branch, newContent, "Update CHANGELOG.md with new version")
	if err != nil {
		return err
	}

	if protectedBranch {
		err = branch_ops.RestoreBranchProtection(glc, projectID, branch)
		if err != nil {
			return err
		}
	}

	if protectedAll {
		err = branch_ops.RestoreBranchProtection(glc, projectID, "*")
		if err != nil {
			return err
		}
	}

	return nil
}

func getChangelogContent(glc *gitlab.Client, projectID, branch string) (string, error) {
	file, _, err := glc.RepositoryFiles.GetFile(projectID, "CHANGELOG.md", &gitlab.GetFileOptions{Ref: &branch})
	if err != nil {
		return "", fmt.Errorf("не удалось скачать CHANGELOG.md из репозитория -> %w", err)
	}

	content, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return "", fmt.Errorf("не удалось получить содержание CHANGELOG.md -> %w", err)
	}

	return string(content), nil
}

func getLastVersion(content string) string {
	re := regexp.MustCompile(`## .*?(\d+\.\d+\.\d+)`)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if matches != nil {
			return matches[1]
		}
	}

	return ""
}

func generateNewVersion(lastVersion, updateFullVersion, updateBuildVersion string) (string, error) {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(lastVersion)
	if matches == nil {
		return "", fmt.Errorf("последняя версия указана в некоректном формате -> %s", lastVersion)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	build, _ := strconv.Atoi(matches[3])

	switch updateFullVersion {
	case "":
		return fmt.Sprintf("%d.%d.%s", major, minor, updateBuildVersion), nil
	default:
		newVersionParts := strings.Split(updateFullVersion, ".")

		if len(newVersionParts) == 1 {
			newBuild, err := strconv.Atoi(newVersionParts[0])
			if err != nil {
				return "", fmt.Errorf("версия указана в некорректном формате -> %s", updateFullVersion)
			}

			if newBuild <= build {
				return "", fmt.Errorf("новая версия должна быть больше предыдущей версии -> %s", lastVersion)
			}

			return fmt.Sprintf("%d.%d.%d", major, minor, newBuild), nil
		}

		if len(newVersionParts) == 3 {
			newMajor, majorErr := strconv.Atoi(newVersionParts[0])
			newMinor, minorErr := strconv.Atoi(newVersionParts[1])
			newBuild, buildErr := strconv.Atoi(newVersionParts[2])
			if majorErr != nil || minorErr != nil || buildErr != nil {
				return "", fmt.Errorf("версия указано в некорректном формате -> %s", updateFullVersion)
			}

			if newMajor < major || (newMajor == major && newMinor < minor) || (newMajor == major && newMinor == minor && newBuild <= build) {
				return "", fmt.Errorf("новая версия должна быть больше предыдущей версии -> %s", lastVersion)
			}

			return updateFullVersion, nil
		}

		return "", fmt.Errorf("версия указана в некорректном формате -> %s", updateFullVersion)
	}
}

func insertNewVersion(content, newVersion, comment string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var header string

	if scanner.Scan() {
		header = scanner.Text()
	}

	newEntry := fmt.Sprintf("## Версия %s\n%s", newVersion, comment)

	if header != "" {
		remainingContent := strings.TrimSpace(content[len(header):])
		return fmt.Sprintf("%s\n\n%s\n\n%s", header, newEntry, remainingContent)
	}

	return fmt.Sprintf("%s\n\n%s%s", header, newEntry, content)
}

func updateChangelogContent(glc *gitlab.Client, projectID, branch, content, commitMessage string) error {
	fileUpdateAction := gitlab.FileUpdate
	filePath := "CHANGELOG.md"

	_, _, err := glc.Commits.CreateCommit(projectID, &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: &commitMessage,
		Actions: []*gitlab.CommitActionOptions{
			{
				Action:   &fileUpdateAction,
				FilePath: &filePath,
				Content:  &content,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("не удалось обновить CHANGELOG.md -> %w", err)
	}

	return nil
}
