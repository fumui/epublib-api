package util

import "regexp"

func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`).MatchString(email)
}
