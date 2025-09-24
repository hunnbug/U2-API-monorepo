package valueObjects

import (
	errs "anketas-service/errors"
	"fmt"
	"regexp"
)

type Username struct {
	Value string
}

func NewUsername(value string) (Username, error) {
	if isValidUsername(value) {
		username := fmt.Sprintf("@%s", value)
		return Username{username}, nil
	} else {
		return Username{}, errs.ErrInvalidLogin
	}
}

func isValidUsername(value string) bool {
	if len(value) < 4 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, value)
	return matched
}
