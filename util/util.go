package util

import "os"

// StrInSlice return a bool value to indicate whether the `str` was in
// the `targets`
func StrInSlice(str string, targets []string) (in bool) {
	for _, t := range targets {
		if str == t {
			in = true
			break
		}
	}
	return
}

// IsFileExist return a bool value to indicate the `file` is exist or not
func IsFileExist(file string) bool {
	_, err := os.Stat(file)
	return !os.IsNotExist(err)
}
