package utils

import (
	"fmt"
	"lisk/logger"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/AlecAivazis/survey/v2"
)

func SetConsoleTitle(title string) error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "title", title)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("не удалось установить название консоли на Windows: %w", err)
		}
	case "darwin":
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Terminal" to set custom title of window 1 to "%s"`, title))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("не удалось установить название консоли через osascript: %w", err)
		}

	default:
		return fmt.Errorf("установка названия консоли не поддерживается на OS: %s", runtime.GOOS)
	}
	return nil
}

func UserChoice() string {
	mainMenu := []string{
		"1. Oku",
		"2. Ionic",
		"3. Relay",
		"4. Portal",
		"0. Exit",
	}

	subMenus := map[string][]string{
		"Ionic": {
			"1. Ionic71Supply",
			"2. Ionic71Borrow",
			"3. IonicWithdraw",
			"0. Back",
		},
		"Portal": {
			"1. Checker",
			"2. Portal_daily_check",
			"3. Portal_main_tasks",
			"0. Back",
		},
	}

	var rgx = regexp.MustCompile(`^\d+\.\s*`)

	for {
		selected := promptSelection("Choose module:", mainMenu)
		selected = rgx.ReplaceAllString(selected, "")

		switch selected {
		case "Oku", "Relay":
			return selected
		case "Ionic", "Portal":
			if subSelected := handleSubMenu(selected, subMenus[selected], rgx); subSelected != "" {
				return subSelected
			}
		case "Exit":
			logger.GlobalLogger.Infof("Exiting program.")
			return ""
		default:
			logger.GlobalLogger.Warnf("Invalid selection: %s", selected)
		}
	}
}

func promptSelection(message string, options []string) string {
	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: message,
		Options: options,
		Default: options[len(options)-1],
	}, &selected); err != nil {
		logger.GlobalLogger.Errorf("Error selecting option: %v", err)
		return ""
	}
	return selected
}

func handleSubMenu(menuName string, subMenu []string, rgx *regexp.Regexp) string {
	for {
		selected := promptSelection("Choose "+menuName+" sub-module:", subMenu)
		selected = rgx.ReplaceAllString(selected, "")

		if selected == "Back" {
			return ""
		}
		return selected
	}
}
