package apkg

import (
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
)

func File_exists(fpath string) bool {
	info, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func File_size(fpath string) (int64, error) {
	info, err := os.Stat(fpath)
	if err != nil {
		if os.IsNotExist(err) {
			return -2, err
		}
		return -1, err
	}
	return info.Size(), nil
}

func Select(applist []App) (App, uint8) {
	menu_items := make([]string, len(applist))
	for idx, app := range applist {
		menu_items[idx] = app.Name
	}
	var prompt = &survey.Select{
		Message: "Select an app:",
		Options: menu_items,
		Description: func(value string, index int) string {
			return "\n    " + string(applist[index].Package)
		},
	}
	selection_idx := 0
	var opts survey.AskOpt = func(options *survey.AskOptions) error {
		options.PromptConfig.Icons.Question.Text = "?"
		options.PromptConfig.Icons.Question.Format = "default"
		options.PromptConfig.Icons.SelectFocus.Format = "default"
		return nil
	}
	if err := survey.AskOne(prompt, &selection_idx, opts); err != nil {
		var (
			app  App
			sink core.OptionAnswer
			val  interface{} = sink
		)
		prompt.Cleanup(nil, val)
		if err == terminal.InterruptErr {
			return app, 130
		}
		return app, 80
	}
	return applist[selection_idx], 0
}

func Confirm(message string, default_ans bool) bool {
	answer := default_ans
	prompt := &survey.Confirm{
		Message: message, Default: default_ans,
	}
	var opts survey.AskOpt = func(options *survey.AskOptions) error {
		options.PromptConfig.Icons.Question.Text = ">"
		options.PromptConfig.Icons.Question.Format = "default"
		return nil
	}
	survey.AskOne(prompt, &answer, opts)
	return answer
}

func Valid_apk(apk string) bool {
	output, err := exec.Command(Aapt_path, "dump", "badging", apk).CombinedOutput()
	result := true
	if err != nil && strings.Contains(string(output), "assets could not be loaded") {
		result = false
	}
	return result
}
