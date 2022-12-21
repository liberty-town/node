package advanced_strings

import "strings"

var (
	SeparatorsDict map[rune]int
)

func Splitter(s string, splits string) []string {
	m := make(map[rune]int)
	for _, r := range splits {
		m[r] = 1
	}

	return strings.FieldsFunc(s, func(r rune) bool {
		return m[r] == 1
	})
}

func SplitterSeparators(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return SeparatorsDict[r] == 1
	})
}

func init() {
	SeparatorsDict = make(map[rune]int)
	for _, r := range " !@#$%^&*()-=[]\\|'\",./<>_+`~	{}:?" {
		SeparatorsDict[r] = 1
	}
}
