package finder

import (
	"slices"
	"strings"
)

// sortBuilder returns a string containing confirmed sorts only.
func sortBuilder(tableAlias, urlSorts string, columns *[]string) string {
	var sortsString string

	sortsString = strings.ReplaceAll(urlSorts, " ", "")
	sortsString = strings.ReplaceAll(sortsString, ";", "")
	sortsString = strings.ReplaceAll(sortsString, "\\", "")

	sortsArr := strings.Split(sortsString, ",")

	var verifiedSortsArr []string

	// In case no sort is provided it defaults to -created_at or -id
	if len(sortsArr) == 1 && sortsArr[0] == "" {
		var orderBy string
		confirmCreatedAt := slices.Contains(*columns, "created_at")
		if confirmCreatedAt {
			orderBy = tableAlias + ".created_at DESC"
		} else {
			orderBy = tableAlias + ".id DESC"
		}
		verifiedSortsArr = append(verifiedSortsArr, orderBy)
	}

	for _, value := range sortsArr {

		var orderBy string
		col := strings.ToLower(value)

		if strings.HasPrefix(value, "-") {
			col = strings.TrimPrefix(value, "-")
			orderBy = tableAlias + "." + col + " DESC"
		} else {
			orderBy = tableAlias + "." + col + " ASC"
		}
		pass := slices.Contains(*columns, col)

		if pass {
			verifiedSortsArr = append(verifiedSortsArr, orderBy)
		}
	}
	sortsString = strings.Join(verifiedSortsArr, ",")

	return sortsString
}
