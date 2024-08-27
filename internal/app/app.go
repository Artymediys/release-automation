package app

import (
	"fmt"
	"os"

	"arel/config"
	"arel/internal/cli"

	"github.com/charmbracelet/huh"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

// glpat-kCRtN7eT8oLzVb3EsdU6
func Run() {
	///////////////////////////////////
	////////// CONFIGURATION //////////
	///////////////////////////////////
	err := config.Read()
	if err != nil {
		var (
			gitlabURL string
			gitlabPAT string
		)

		err = huh.NewForm(cli.AskForConfig(&gitlabURL, &gitlabPAT)).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, cli.ErrorForm, err)
			return
		}

		err = config.Create(gitlabURL, gitlabPAT)
		if err != nil {
			fmt.Fprintln(os.Stderr, "не удалось создать конфиг -> ", err)
			return
		}
	}

	glc, err := gitlab.NewClient(viper.GetString("pat"), gitlab.WithBaseURL(viper.GetString("url")))
	if err != nil {
		fmt.Fprintln(os.Stderr, "не удалось создать клиент для взаимодействия с GitLab API -> ", err)
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
		version      string
		comment      string
	)

	err = cli.QA(glc, &projectIDs, &projectNames, &group, &sourceBranch, &targetBranch, &version, &comment)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for i := 0; i < len(projectIDs); i++ {
		err = cli.Action(glc, projectIDs[i], projectNames[i], sourceBranch, targetBranch, version, comment)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}
