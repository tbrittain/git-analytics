package query

import "strings"

// buildExcludeClauses returns a SQL fragment like " AND col NOT GLOB ? AND col NOT GLOB ?"
// and the corresponding args slice. Returns ("", nil) when globs is empty.
func buildExcludeClauses(column string, globs []string) (string, []any) {
	if len(globs) == 0 {
		return "", nil
	}
	var b strings.Builder
	args := make([]any, len(globs))
	for i, g := range globs {
		b.WriteString(" AND ")
		b.WriteString(column)
		b.WriteString(" NOT GLOB ?")
		args[i] = g
	}
	return b.String(), args
}
