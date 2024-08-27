package version_ops

import (
	"fmt"

	"github.com/xanzy/go-gitlab"
)

func CreateTag(glc *gitlab.Client, projectID, branch, version string) error {
	var tagName string
	var tagMessage = fmt.Sprintf("Tag %s created from branch %s", tagName, branch)

	switch branch {
	case "rc":
		tagName = fmt.Sprintf("%s-rc1", version)
	default:
		tagName = version
	}

	opts := &gitlab.CreateTagOptions{
		TagName: &tagName,
		Ref:     &branch,
		Message: &tagMessage,
	}

	_, _, err := glc.Tags.CreateTag(projectID, opts)
	if err != nil {
		return fmt.Errorf("не удалось создать тег -> %w", err)
	}

	return nil
}
