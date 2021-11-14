package regexpchecker

import (
	"fmt"
	"regexp"
)

//FindReMatch ... Returns the string matched and a bool indicating if the match was successfull
func FindReMatch(someText string) (string, bool) {
	pattern := `(col-)([\w]*)(-srl-[0-9-]*)`
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		panic(fmt.Sprintf("stripOutInstitution error: %s", err.Error()))
	}
	matched, err := regexp.MatchString(pattern, someText)
	if err != nil {
		panic(err)
	}
	if matched {
		stringFound := matcher.FindSubmatch([]byte(someText))[2]
		return fmt.Sprintf("%s\n", stringFound), matched
	}
	return "notfound", matched
}
