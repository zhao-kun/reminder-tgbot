package util

import "os"

// StrInSlice return a bool type value to indicated whether the `str` was in
// the `target`
func StrInSlice(str string, targets []string) (in bool) {
	for _, t := range targets {
		if str == t {
			in = true
			break
		}
	}
	return
}

// IsFileExist return a bool to indicated file is exist or not
func IsFileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
