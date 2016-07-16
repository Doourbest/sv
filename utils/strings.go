package utils

import(
	"strings"
)

func UcFirst(s string) string{
	if len(s)==0 {
		return ""
	}
	return strings.ToUpper(s[0:1])+s[1:];
}

func LcFirst(s string) string{
	if len(s)==0 {
		return ""
	}
	return strings.ToLower(s[0:1])+s[1:];
}

