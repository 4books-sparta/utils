package utils

import "regexp"

func IsSlug(in string) bool {
	reg := regexp.MustCompile("^([a-z0-9_-]{1,500})$")
	valid := reg.MatchString(in)
	return valid
}
