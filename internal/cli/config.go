package cli

import "github.com/charmbracelet/huh"

func AskForConfig(gitlabURL, gitlabPAT *string) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Введите GitLab URL").
			Placeholder("https://gitlab.example.com").
			Prompt("URL:").
			Value(gitlabURL),
		huh.NewInput().
			Title("Введите GitLab PAT").
			Placeholder("Personal Access Token").
			Prompt("PAT:").
			Value(gitlabPAT),
	)
}
