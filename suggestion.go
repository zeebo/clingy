package clingy

import (
	"fmt"
	"strings"

	"github.com/zeebo/errs/v2"
)

func (env *Environment) appendUnknownCommandErrorWithSuggestions(st *runState, descs []cmdDesc) {
	if env.DisableSuggestions {
		env.appendUnknownCommandError(st)
		return
	}

	if env.SuggestionsMinEditDistance <= 0 {
		env.SuggestionsMinEditDistance = 2
	}

	name, ok, err := st.peekName()
	if ok {
		suggestionsString := fmt.Sprintf("%q for %q", name, st.name())
		if suggestions := suggestionsFor(name, descs, env.SuggestionsMinEditDistance); len(suggestions) > 0 {
			suggestionsString += "\n\n\tMaybe you meant:\n"
			for _, s := range suggestions {
				suggestionsString += fmt.Sprintf("\t\t%v\n", s)
			}
		}

		st.errors = append(st.errors, errs.Tag("unknown command").Errorf("%s", suggestionsString))
	}
	if err != nil {
		st.errors = append(st.errors, err)
	}
}

func suggestionsFor(typedCmd string, cmds []cmdDesc, distance int) []string {
	suggestions := []string{}
	for _, cmd := range cmds {
		levenshteinDistance := levenshteinDistance(typedCmd, cmd.name)
		suggestByLevenshtein := levenshteinDistance <= distance
		suggestByPrefix := strings.HasPrefix(strings.ToLower(cmd.name), strings.ToLower(typedCmd))
		if suggestByLevenshtein || suggestByPrefix {
			suggestions = append(suggestions, cmd.name)
		}
	}
	return suggestions
}

// levenshteinDistance compares two strings and returns the Levenshtein Distance between them.
func levenshteinDistance(a, b string) int {
	a = strings.ToLower(a)
	b = strings.ToLower(b)

	d := make([][]int, len(a)+1)
	for i := range d {
		d[i] = make([]int, len(b)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(b); j++ {
		for i := 1; i <= len(a); i++ {
			if a[i-1] == b[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				min := d[i-1][j]
				if d[i][j-1] < min {
					min = d[i][j-1]
				}
				if d[i-1][j-1] < min {
					min = d[i-1][j-1]
				}
				d[i][j] = min + 1
			}
		}
	}
	return d[len(a)][len(b)]
}
