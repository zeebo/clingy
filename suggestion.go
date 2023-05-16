package clingy

import (
	"fmt"
	"strings"

	"github.com/zeebo/errs/v2"
)

func (env *Environment) appendUnknownCommandErrorWithSuggestions(st *runState, descs []cmdDesc) {
	dist := env.SuggestionsMinEditDistance
	if dist < 0 {
		env.appendUnknownCommandError(st)
		return
	} else if dist == 0 {
		dist = 2
	}

	name, ok, err := st.peekName()
	if ok {
		var sbuild strings.Builder
		fmt.Fprintf(&sbuild, "%q", name)
		if suggestions := suggestionsFor(name, descs, dist); len(suggestions) > 0 {
			sbuild.WriteString(". did you mean:")
			for _, s := range suggestions {
				sbuild.WriteString("\n\t\t")
				sbuild.WriteString(s)
			}
		}

		st.errors = append(st.errors, errs.Tag("unknown command").Errorf("%s", sbuild.String()))
	}
	if err != nil {
		st.errors = append(st.errors, err)
	}
}

func (env *Environment) appendUnknownCommandError(st *runState) {
	name, ok, err := st.peekName()
	if ok {
		st.errors = append(st.errors, errs.Tag("unknown command").Errorf("%q", name))
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
