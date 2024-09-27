package cli

import (
	"fmt"
	"log"
	"time"

	"arel/internal/gitlab/branch_ops"
	"arel/internal/gitlab/repo_ops"
	"arel/internal/gitlab/version_ops"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/xanzy/go-gitlab"
)

func QA(
	glc *gitlab.Client,
	projectIDs, projectNames *[]string,
	group, sourceBranch, targetBranch, fullVersion, buildVersion, comment *string,
) error {
	var (
		groupForm *huh.Group
		groupErr  error
	)

	////////////////////////////////////
	///////// ГРУППЫ И ПРОЕКТЫ /////////
	////////////////////////////////////
	log.Println("получаем данные о проектах...")
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
	log.Println("получаем данные о ветках проектов...")
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
	log.Println("получаем данные о новой версии...")
	err = huh.NewForm(AskForVersion(fullVersion, comment)).Run()
	if err != nil {
		return fmt.Errorf(ErrorForm+"%w", err)
	}

	if *fullVersion == "" {
		*buildVersion = time.Now().Format("0601021504")
	}

	/////////////////////////////////////
	/////////// ПОДТВЕРЖДЕНИЕ ///////////
	/////////////////////////////////////
	//log.Println("утверждаем выбор пользователя...")
	//err = huh.NewForm()

	return nil
}

func Action(
	glc *gitlab.Client,
	projectID, projectName, sourceBranch, targetBranch, comment, fullVersion, buildVersion string,
) (Stage, error) {
	var (
		gitlabErr error
		spinErr   error

		stageStatus Stage
	)

	///////////////////////////////////
	///// ОБНОВЛЕНИЕ CHANGELOG.md /////
	///////////////////////////////////
	log.Println(fmt.Sprintf("обновляем CHANGELOG.md в \"%s\"...", projectName))
	spinErr = spinner.New().
		Title(fmt.Sprintf("Обновляем CHANGELOG.md в \"%s\"...", projectName)).
		Action(func() {
			gitlabErr = version_ops.CheckAndUpdateVersion(glc, projectID, sourceBranch, comment, buildVersion, &fullVersion)
		}).Run()
	if spinErr != nil {
		stageStatus.Changelog = -1
		return stageStatus, fmt.Errorf(ErrorSpinner+"%w", spinErr)
	}
	if gitlabErr != nil {
		stageStatus.Changelog = -1
		return stageStatus, gitlabErr
	}

	stageStatus.Changelog = 1

	////////////////////////////////////
	////////// СЛИЯЕНИЕ ВЕТОК //////////
	////////////////////////////////////
	log.Println(fmt.Sprintf("сливаем ветки \"%s\" -> \"%s\"...", sourceBranch, targetBranch))
	spinErr = spinner.New().
		Title(fmt.Sprintf("Сливаем ветки \"%s\" -> \"%s\"...", sourceBranch, targetBranch)).
		Action(func() {
			if (targetBranch == "main" || targetBranch == "master") && targetBranch != sourceBranch {
				gitlabErr = branch_ops.MergeBranches(glc, projectID, sourceBranch, targetBranch)
			} else {
				gitlabErr = branch_ops.ForcePushBranch(glc, projectID, sourceBranch, targetBranch)
			}
		}).Run()
	if spinErr != nil {
		stageStatus.MergePush = -1
		return stageStatus, fmt.Errorf(ErrorSpinner+"%w", spinErr)
	}
	if gitlabErr != nil {
		stageStatus.MergePush = -1
		return stageStatus, gitlabErr
	}

	stageStatus.MergePush = 1

	///////////////////////////////////
	////////// СОЗДАНИЕ ТЕГА //////////
	///////////////////////////////////
	log.Println("формируем тег...")
	if targetBranch != "main" && targetBranch != "master" {
		spinErr = spinner.New().
			Title("Формируем тег...").
			Action(func() {
				gitlabErr = version_ops.CreateTag(glc, projectID, targetBranch, fullVersion)
			}).Run()
		if spinErr != nil {
			stageStatus.Tag = -1
			return stageStatus, fmt.Errorf(ErrorSpinner+"%w", spinErr)
		}
		if gitlabErr != nil {
			stageStatus.Tag = -1
			return stageStatus, gitlabErr
		}

		stageStatus.Tag = 1
	}

	return stageStatus, nil
}
