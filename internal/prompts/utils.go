package prompts

import (
	"github.com/erikgeiser/promptkit/confirmation"
)

// confirm asks the given yes/no question with the standard confirmation
// prompt and reports the user's answer.
func confirm(question string) (bool, error) {
	c := confirmation.New(question, confirmation.Undecided)
	return c.RunPrompt()
}
