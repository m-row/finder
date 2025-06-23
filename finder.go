package finder

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/Masterminds/squirrel"
)

type filterOp struct {
	input    string
	criteria string
	op       string
	values   []string
}

func find(c *configFinder) error { //nolint: gocyclo,maintidx // unavoidable
	tableName := c.Model.TableName()
	page, err := strconv.ParseUint(
		strings.TrimSpace(c.UrlValues.Get("page")),
		10,
		64,
	)
	if err != nil {
		page = 1
	}
	paginate, err := strconv.ParseUint(
		strings.TrimSpace(c.UrlValues.Get("paginate")),
		10,
		64,
	)
	if err != nil {
		paginate = 12
	}
	// assign url query to defined struct
	c.UrlQuery = &URLQuery{
		Q:        strings.TrimSpace(c.UrlValues.Get("q")),
		Page:     page,
		Sorts:    strings.TrimSpace(c.UrlValues.Get("sorts")),
		Filters:  strings.TrimSpace(c.UrlValues.Get("filters")),
		Paginate: paginate,
	}

	if c.IsPublic &&
		slices.Contains(*c.Meta.Columns, "is_disabled") {
		p := fmt.Sprintf("%s.is_disabled=false", tableName)
		*c.Results = c.Results.Where(p)
	}

	if c.UrlQuery.Filters != "" {
		filters := filterer(
			c.Results,
			c.UrlQuery.Filters,
			tableName,
			c.Meta.Columns,
			c.GroupBys,
			c.Model.Relations(),
		)

		for _, v := range filters {
			switch v.op {
			case "=":
				*c.Results = c.Results.Where(squirrel.Eq{v.criteria: v.values})
			case "!=":
				*c.Results = c.Results.
					Where(squirrel.NotEq{v.criteria: v.values})
			case ">":
				if len(v.values) == 1 {
					*c.Results = c.Results.
						Where(squirrel.Gt{v.criteria: v.values[0]})
				}
			case ">=":
				if len(v.values) == 1 {
					*c.Results = c.Results.
						Where(squirrel.GtOrEq{v.criteria: v.values[0]})
				}
			case "<":
				if len(v.values) == 1 {
					*c.Results = c.Results.
						Where(squirrel.Lt{v.criteria: v.values[0]})
				}
			case "<=":
				if len(v.values) == 1 {
					*c.Results = c.Results.
						Where(squirrel.LtOrEq{v.criteria: v.values[0]})
				}
			case "BETWEEN":
				if len(v.values) == 2 {
					*c.Results = c.Results.
						Where(squirrel.GtOrEq{v.criteria: v.values[0]}).
						Where(squirrel.LtOrEq{v.criteria: v.values[1]})
				}
			case "na":
				*c.Results = c.Results.Where(squirrel.Eq{v.criteria: nil})
			case "nn":
				*c.Results = c.Results.Where(squirrel.NotEq{v.criteria: nil})
			default:
				*c.Results = c.Results.Where(squirrel.Eq{v.criteria: v.values})
			}
		}
	}

	if c.UrlQuery.Q != "" && c.Meta.SearchColumns != nil {

		// TODO: write better search
		var searchSQL []squirrel.Sqlizer

		for _, col := range *c.Meta.SearchColumns {
			switch {
			case col == "fulltext":
				fulltextQArgs := strings.Fields(c.UrlQuery.Q)
				fullText := strings.Join(fulltextQArgs, "+") + ":*"
				sql := fmt.Sprintf(
					"%s.fulltext @@  websearch_to_tsquery('arabic', ?)",
					tableName,
				)
				searchSQL = append(searchSQL, squirrel.Expr(sql, fullText))
			case strings.HasPrefix(col, "rel:"):
				// use in model as SearchFields array as:
				//  "rel:customers.name" use aliases of the join
				searchCol, found := strings.CutPrefix(
					col,
					"rel:",
				) // TODO: add a check to relation columns
				if found {
					searchVal := fmt.Sprint("%", c.UrlQuery.Q, "%")
					searchSQL = append(
						searchSQL,
						squirrel.ILike{searchCol: searchVal},
					)
				}
			default:
				searchCol := fmt.Sprintf("%s.%s", tableName, col)
				searchVal := fmt.Sprint("%", c.UrlQuery.Q, "%")
				searchSQL = append(
					searchSQL,
					squirrel.ILike{searchCol: searchVal},
				)
			}
		}

		finalSearchQuery := "("
		var searchArgs []any

		for index, v := range searchSQL {
			query, args, err := v.ToSql()
			if err == nil {
				searchArgs = append(searchArgs, args...)
				finalSearchQuery += query

				if index < len(searchSQL)-1 {
					finalSearchQuery += " OR "
				} else {
					finalSearchQuery += ")"
				}
			}
		}
		*c.Results = c.Results.Where(finalSearchQuery, searchArgs...)
	}

	// from / to queries -------------- ---------------------------------------
	hasFrom := c.UrlValues.Has("from")
	if hasFrom {
		timezone := "UTC"
		if c.Timezone != nil {
			timezone = *c.Timezone
		}
		loc, err := time.LoadLocation(timezone)
		if err != nil {
			return err
		}
		fromVal := c.UrlValues.Get("from")
		from, err := time.Parse(time.RFC3339, fromVal)
		if err != nil {
			fromDate, err := time.Parse(time.DateOnly, fromVal)
			if err != nil {
				return err
			}
			from = fromDate
		}
		fromToCol := "created_at"
		if c.FromToCol == nil {
			c.FromToCol = &fromToCol
		}

		if slices.Contains(*c.Meta.Columns, *c.FromToCol) {

			year, month, day := time.Now().Date()
			to := time.Date(year, month, day, 23, 59, 59, 0, loc)

			if hasFrom && c.UrlValues.Has("to") {
				toVal := c.UrlValues.Get("to")
				toParsed, err := time.Parse(
					time.RFC3339,
					toVal,
				)
				if err != nil {
					toDate, err := time.Parse(time.DateOnly, toVal)
					if err != nil {
						// TODO: test this
						to = time.Date(year, month, day, 23, 59, 59, 0, loc)
					} else {
						to = toDate
					}
				} else {
					to = toParsed
				}
			}

			*c.Results = c.Results.Where(
				"("+
					tableName+"."+*c.FromToCol+" AT TIME ZONE '"+timezone+"', "+
					tableName+"."+*c.FromToCol+" AT TIME ZONE '"+timezone+
					"') OVERLAPS (?, ?)",
				from.Format(time.DateTime),
				to.Format(time.DateTime),
			)
		}
	}

	sorts := sortBuilder(
		tableName,
		c.UrlQuery.Sorts,
		c.IDColumn,
		c.Meta.Columns,
	)
	if sorts != "" {
		*c.Results = c.Results.OrderByClause(sorts)
	}

	if c.OverrideSort != "" {
		*c.Results = c.Results.OrderByClause(c.OverrideSort)
	}

	if c.GroupBys != nil {
		if len(*c.GroupBys) > 0 {
			*c.Results = c.Results.
				GroupBy(*c.GroupBys...)
		}
	}
	return nil
}
