package utils

import (
	"fmt"
	"os/exec"
)

func RunGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка при выполнении команды git: %w\nВывод: %s", err, string(output))
	}

	return nil
}
