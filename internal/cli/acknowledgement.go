package cli

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

func AskForAcknowledgement(
	confirm *bool,
	projectNames *[]string,
	group, sourceBranch, targetBranch, fullVersion, comment *string,
) *huh.Group {
	var (
		versionMessage      string
		commentMessage      string
		projectNamesMessage string
	)

	if *fullVersion == "" {
		versionMessage = "не была указана, поэтому будет сформирована автоматически"
	}

	if *comment == "" {
		commentMessage = "не был указан, поэтому будет сформирован автоматически"
	}

	for _, projectName := range *projectNames {
		projectNamesMessage += fmt.Sprintln(projectName)
	}

	ackText := fmt.Sprintf(
		"ПОДТВЕРДИТЕ СВОЙ ВЫБОР\nГруппа: %s\nВетки: %s -> %s\nВерсия: %s\nКомментарий: %s\nПроекты:\n%s",
		*group, *sourceBranch, *targetBranch, versionMessage, commentMessage, projectNamesMessage,
	)

	return huh.NewGroup(
		huh.NewConfirm().
			Title(ackText).
			Affirmative("Да, Let's Go!").
			Negative("Нет, Отменяем!").
			Value(confirm),
	)
}
