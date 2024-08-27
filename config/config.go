package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func Create(url, pat string) error {
	err := setConfigPath()
	if err != nil {
		return fmt.Errorf("не удаётся получить домашнюю директорию пользователя -> %w", err)
	}

	viper.Set("url", url)
	viper.Set("pat", pat)

	if err = viper.WriteConfig(); err != nil {
		return fmt.Errorf("не удаётся записать конфигурационный файл -> %w", err)
	}

	return nil
}

func Read() error {
	err := setConfigPath()
	if err != nil {
		return fmt.Errorf("не удаётся получить домашнюю директорию пользователя -> %w", err)
	}

	if err = viper.ReadInConfig(); err != nil {
		return fmt.Errorf("не удаётся прочитать конфигурационный файл -> %w", err)
	}

	if !viper.IsSet("url") {
		return fmt.Errorf("не удаётся найти GitLab URL")
	}

	if !viper.IsSet("pat") {
		return fmt.Errorf("не удаётся найти GitLab Personal Access Token")
	}

	return nil
}

func setConfigPath() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	viper.SetConfigFile(home + "/.arel.yaml")

	return nil
}
