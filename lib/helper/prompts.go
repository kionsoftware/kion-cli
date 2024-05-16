package helper

import "github.com/AlecAivazis/survey/v2"

////////////////////////////////////////////////////////////////////////////////
//                                                                            //
//  Prompts                                                                   //
//                                                                            //
////////////////////////////////////////////////////////////////////////////////

// surveyFormat sets survey icon and color configs.
var surveyFormat = survey.WithIcons(func(icons *survey.IconSet) {
	icons.Question.Text = ""
	icons.Question.Format = "default+hb"
})

// PromptSelect prompts the user to select from a slice of options. It requires
// that the selection made be one of the options provided.
func PromptSelect(message string, options []string) (string, error) {
	selection := ""
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}
	err := survey.AskOne(prompt, &selection, surveyFormat)
	return selection, err
}

// PromptInput prompts the user to provide dynamic input.
func PromptInput(message string) (string, error) {
	var input string
	pi := &survey.Input{
		Message: message,
	}
	err := survey.AskOne(pi, &input, surveyFormat, survey.WithValidator(survey.Required))
	return input, err
}

// PromptPassword prompts the user to provide sensitive dynamic input.
func PromptPassword(message string) (string, error) {
	var input string
	pi := &survey.Password{
		Message: message,
	}
	err := survey.AskOne(pi, &input, surveyFormat, survey.WithValidator(survey.Required))
	return input, err
}
