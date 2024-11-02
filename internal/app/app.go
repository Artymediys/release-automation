package app

import (
	"fmt"
	"log"
	"strings"
	"time"

	"arel/config"
	"arel/internal/cli"
	"arel/internal/report"
	"arel/pkg/utils"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

func Run() {
	///////////////////////////////////
	//////// LOGGER / REPORTER ////////
	///////////////////////////////////
	logFile, err := utils.NewLogger()
	if err != nil {
		log.Println("не удалось настроить логирование -> ", err)
		return
	}
	defer logFile.Close()

	reportFile, err := report.NewReporter()
	if err != nil {
		log.Println("не удалось настроить репортера -> ", err)
		return
	}
	defer reportFile.Close()

	///////////////////////////////////
	////////// CONFIGURATION //////////
	///////////////////////////////////
	err = config.Read()
	if err != nil {
		var (
			gitlabURL string
			gitlabPAT string
		)

		err = huh.NewForm(cli.AskForConfig(&gitlabURL, &gitlabPAT)).WithTheme(huh.ThemeBase()).Run()
		if err != nil {
			log.Println(cli.ErrorForm, err)
			return
		}

		err = config.Create(gitlabURL, gitlabPAT)
		if err != nil {
			log.Println("не удалось создать конфиг -> ", err)
			return
		}
	}

	glc, err := gitlab.NewClient(viper.GetString("pat"), gitlab.WithBaseURL(viper.GetString("url")))
	if err != nil {
		log.Println("не удалось создать клиент для взаимодействия с GitLab API -> ", err)
		return
	}

	///////////////////////////////////
	/////////// APPLICATION ///////////
	///////////////////////////////////
	var (
		group        string
		projectIDs   []string
		projectNames []string
		sourceBranch string
		targetBranch string
		fullVersion  string
		buildVersion string
		comment      string
		confirm      bool
	)

	err = cli.QA(glc, &projectIDs, &projectNames, &group, &sourceBranch, &targetBranch, &fullVersion, &buildVersion, &comment, &confirm)
	if err != nil {
		log.Println(err)
		return
	}

	var resultString string
	for i := 0; i < len(projectIDs); i++ {
		projectName := strings.Join(strings.Fields(projectNames[i]), " ")

		appStage, err := cli.Action(glc, projectIDs[i], projectName, sourceBranch, targetBranch, comment, fullVersion, buildVersion)
		if err != nil {
			log.Println(err)
		}

		resultString += fmt.Sprintf(
			"Рапорт от: %s\nПроект: %s\nИзменение CHANGELOD.md - %s\nMerge / Force Push – %s\nСоздание тега – %s\n\n",
			time.Now().Format(time.DateTime), projectName,
			appStage.Changelog.Status(), appStage.MergePush.Status(), appStage.Tag.Status(),
		)
		_, err = reportFile.WriteString(resultString)
		if err != nil {
			log.Println("не удалось записать рапорт в файл ->", err)
		}

		log.Println("проверяем наличие оставшихся проектов...")
		spinErr := spinner.New().
			Title("Проверяем наличие оставшихся проектов...").
			Action(func() {
				time.Sleep(10 * time.Second)
			}).Run()
		if spinErr != nil {
			log.Println(cli.ErrorSpinner+"%w", spinErr)
		}
	}

	fmt.Println(resultString)

	log.Println("ARel: полёт закончен!")
}
