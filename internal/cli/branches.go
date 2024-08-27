package cli

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/xanzy/go-gitlab"
)

func AskForBranches(
	glc *gitlab.Client,
	projectIDs *[]string,
	projectNames *[]string,
	sourceBranch *string,
	targetBranch *string,
	getProjectID func(*gitlab.Client, string) (string, error),
	getCommonBranches func(*gitlab.Client, []string) ([]string, error),
) (*huh.Group, error) {

	*projectIDs = make([]string, len(*projectNames))

	for i, projectName := range *projectNames {
		projectID, err := getProjectID(glc, projectName)
		if err != nil {
			return nil, fmt.Errorf("не удалось получить ID для проекта %s -> %w", projectName, err)
		}
		(*projectIDs)[i] = projectID
	}

	branches, err := getCommonBranches(glc, *projectIDs)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить общие ветки проектов: %v", err)
	}

	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Выберите исходную ветку").
			Value(sourceBranch).
			Height(len(branches)+2).
			Options(huh.NewOptions(branches...)...),
		huh.NewSelect[string]().
			Title("Выберите целевую ветку").
			Value(targetBranch).
			Height(len(branches)+2).
			Options(huh.NewOptions(branches...)...).
			Validate(func(chosenBranch string) error {
				if chosenBranch == *sourceBranch {
					return fmt.Errorf("исходная ветка %s совпадает с целевой веткой %s", chosenBranch, *sourceBranch)
				}
				return nil
			}),
	), nil
}
