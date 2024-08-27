package cli

import (
	"fmt"

	"arel/internal/gitlab/branch_ops"
	"arel/internal/gitlab/repo_ops"
	"arel/internal/gitlab/version_ops"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/xanzy/go-gitlab"
)

var (
	ErrorSpinner = "возникла ошибка при отображении интерфейса загрузки -> "
	ErrorForm    = "возникла ошибка при отображении пользовательского интерфейса -> "
)

func QA(glc *gitlab.Client, projectIDs, projectNames *[]string, group, sourceBranch, targetBranch, version, comment *string) error {
	var (
		groupForm *huh.Group
		groupErr  error
	)

	////////////////////////////////////
	///////// ГРУППЫ И ПРОЕКТЫ /////////
	////////////////////////////////////
	err := spinner.New().
		Title("Получаем данные о проектах...").
		Action(func() {
			groupForm, groupErr = AskForProjects(glc, group, projectNames, repo_ops.GetGroupNames, repo_ops.GetGroupProjectMap)
		}).Run()
	if err != nil {
		return fmt.Errorf(ErrorSpinner+"%w", err)
	}

	if groupErr != nil {
		return fmt.Errorf("возникла ошибка при формировании интерфейса для групп и проектов -> %w", groupErr)
	}

	err = huh.NewForm(groupForm).Run()
	if err != nil {
		return fmt.Errorf(ErrorForm+"%w", err)
	}

	if len(*projectNames) <= 0 {
		return fmt.Errorf("возникла ошибка при выборе проектов -> должен быть выбран хотя бы 1 проект")
	}

	///////////////////////////////////
	////////////// ВЕТКИ //////////////
	///////////////////////////////////
	err = spinner.New().
		Title("Получаем данные о ветках проектов...").
		Action(func() {
			groupForm, groupErr = AskForBranches(
				glc, projectIDs, projectNames, sourceBranch, targetBranch,
				repo_ops.GetProjectID, branch_ops.GetCommonBranches,
			)
		}).Run()
	if err != nil {
		return fmt.Errorf(ErrorSpinner+"%w", err)
	}

	if groupErr != nil {
		return fmt.Errorf("возникла ошибка при формировании интерфейса для веток проекта -> %w", groupErr)
	}

	err = huh.NewForm(groupForm).Run()
	if err != nil {
		return fmt.Errorf(ErrorForm+"%w", err)
	}

	////////////////////////////////////
	////////////// ВЕРСИЯ //////////////
	////////////////////////////////////
	err = huh.NewForm(AskForVersion(version, comment)).Run()
	if err != nil {
		return fmt.Errorf(ErrorForm+"%w", err)
	}

	return nil
}

func Action(glc *gitlab.Client, projectID, projectName, sourceBranch, targetBranch, version, comment string) error {
	var (
		gitlabErr error
		spinErr   error
	)

	///////////////////////////////////
	///// ОБНОВЛЕНИЕ CHANGELOG.md /////
	///////////////////////////////////
	spinErr = spinner.New().
		Title(fmt.Sprintf("Обновляем CHANGELOG.md в \"%s\"...", projectName)).
		Action(func() {
			gitlabErr = version_ops.CheckAndUpdateVersion(glc, projectID, sourceBranch, comment, &version)
		}).Run()
	if spinErr != nil {
		return fmt.Errorf(ErrorSpinner+"%w", spinErr)
	}
	if gitlabErr != nil {
		return gitlabErr
	}

	////////////////////////////////////
	////////// СЛИЯЕНИЕ ВЕТОК //////////
	////////////////////////////////////
	spinErr = spinner.New().
		Title("Сливаем ветки...").
		Action(func() {
			switch targetBranch {
			case "main", "master":
				gitlabErr = branch_ops.MergeBranches(glc, projectID, sourceBranch, targetBranch)
			default:
				gitlabErr = branch_ops.ForcePushBranch(glc, projectID, sourceBranch, targetBranch)
			}
		}).Run()
	if spinErr != nil {
		return fmt.Errorf(ErrorSpinner+"%w", spinErr)
	}
	if gitlabErr != nil {
		return gitlabErr
	}

	///////////////////////////////////
	////////// СОЗДАНИЕ ТЕГА //////////
	///////////////////////////////////
	if targetBranch != "main" && targetBranch != "master" {
		spinErr = spinner.New().
			Title("Формируем тег...").
			Action(func() {
				gitlabErr = version_ops.CreateTag(glc, projectID, targetBranch, version)
			}).Run()
		if spinErr != nil {
			return fmt.Errorf(ErrorSpinner+"%w", spinErr)
		}
		if gitlabErr != nil {
			return gitlabErr
		}
	}

	return nil
}
