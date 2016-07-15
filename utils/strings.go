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
