package cli

import "github.com/charmbracelet/huh"

func AskForVersion(version, comment *string) *huh.Group {
	return huh.NewGroup(
		huh.NewInput().
			Title("Введите новую версию").
			Prompt("Версия:").
			Placeholder("примеры -> 1.2.3 или 42 или нажмите enter для auto-версии").
			Value(version),
		huh.NewInput().
			Title("Введите комментарий к версии").
			Prompt("Комментарий:").
			Placeholder("Выпустили Арла на свободу.Американский флаг к Арлу не прилагается").
			Value(comment),
	)
}
