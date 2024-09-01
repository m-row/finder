package finder

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/squirrel"
)

// filterer returns confirmed list of filter operations
//
// Example:
// URL Query
//
// localhost?filters=price:gt:40,status:pending|complete,id:555|444
//
// results in an array:
//
//	[
//		{c1.price > [40]}
//		{c1.id = [555 444]}
//		{c1.status = [pending complete]}
//	]
//
// note: gt,gte,lt,lte will only filter on a single value
func filterer(
	results *squirrel.SelectBuilder,
	filters, tableAlias string,
	columns, groupBys *[]string,
	relations *[]RelationField,
) []filterOp {
	var filtersArr []filterOp

	filters = strings.ReplaceAll(filters, ";", "")
	filters = strings.ReplaceAll(filters, "\\", "")

	filtersSplit := strings.Split(filters, ",")

	for _, v := range filtersSplit {
		filterArgs := strings.Split(v, ":")
		if len(filterArgs) > 0 && columns != nil && groupBys != nil {

			//  confirm the first arg as column
			// done separately to avoid removing whitespace from values
			pass := slices.Contains(*columns, filterArgs[0])
			filterArgs[0] = strings.ReplaceAll(filterArgs[0], " ", "")
			partOfRelation := false
			var currentRelation RelationField

			if len(*relations) > 0 {
				for _, rel := range *relations {
					if rel.Table == filterArgs[0] {
						partOfRelation = true
						currentRelation = rel
						if rel.Join == nil {
							filterArgs[0] = rel.Table + ".id"
						} else {
							partOfRelation = true
							currentRelation = rel
							filterArgs[0] = rel.Join.To
							*groupBys = append(*groupBys, filterArgs[0])

							if rel.Through != nil {
								joinString := fmt.Sprintf(
									"%s on %s = %s",
									rel.Table,
									rel.Join.To,
									rel.Through.Join.To,
								)
								throughString := fmt.Sprintf(
									"%s on %s = %s",
									rel.Through.Table,
									rel.Through.Join.From,
									rel.Join.From,
								)
								*results = results.
									LeftJoin(throughString).LeftJoin(joinString)
							} else {
								joinString := fmt.Sprintf(
									"%s on %s = %s",
									rel.Table,
									rel.Join.From,
									rel.Join.To,
								)
								*results = results.LeftJoin(joinString)
							}
						}
					}
				}
			}

			if pass || partOfRelation {
				var currFilter filterOp

				if len(filterArgs) == 2 {
					// is_disabled:false
					// status:completed|cancelled|returned
					filterArgs[0] = strings.ReplaceAll(filterArgs[0], " ", "")

					if partOfRelation {
						currFilter.criteria = filterArgs[0]
					} else {
						currFilter.criteria = tableAlias + "." + filterArgs[0]
					}

					switch filterArgs[1] {
					case "null":
						if currentRelation.Join != nil &&
							currentRelation.Through != nil {
							countString := fmt.Sprintf(
								`(select count(*) from %s where %s = %s) = 0`,
								currentRelation.Through.Table,
								currentRelation.Through.Join.From,
								currentRelation.Join.From,
							)
							*results = results.Where(countString)
						}
						currFilter.op = "na"
						currFilter.input = ""
					case "not-null":
						if currentRelation.Join != nil &&
							currentRelation.Through != nil {
							countString := fmt.Sprintf(
								`(select count(*) from %s where %s = %s) > 0`,
								currentRelation.Through.Table,
								currentRelation.Through.Join.From,
								currentRelation.Join.From,
							)
							*results = results.Where(countString)
						}
						currFilter.op = "nn"
						currFilter.input = ""
					default:
						currFilter.op = "="
						currFilter.input = filterArgs[1]
					}
				}
				if len(filterArgs) == 3 {
					// price:gt:5
					// status:eq:completed|cancelled|returned
					filterArgs[1] = strings.ReplaceAll(filterArgs[1], " ", "")

					switch filterArgs[1] {
					case "eq":
						filterArgs[1] = "="
					case "nq":
						filterArgs[1] = "!="
					case "gt":
						filterArgs[1] = ">"
					case "gte":
						filterArgs[1] = ">="
					case "lt":
						filterArgs[1] = "<"
					case "lte":
						filterArgs[1] = "<="
					case "ex": // exactly - returns row that has 1 of relation
						if currentRelation.Join != nil &&
							currentRelation.Through != nil {
							countString := fmt.Sprintf(
								`(select count(*) from %s where %s = %s) = 1`,
								currentRelation.Through.Table,
								currentRelation.Through.Join.From,
								currentRelation.Join.From,
							)
							*results = results.Where(countString)
						}
						filterArgs[1] = "="
					default:
						filterArgs[1] = "="
					}

					if partOfRelation {
						currFilter.criteria = filterArgs[0]
					} else {
						currFilter.criteria = tableAlias + "." + filterArgs[0]
					}

					currFilter.op = filterArgs[1]
					currFilter.input = filterArgs[2]
				}

				// check if filter has multiple values and make it a slice
				currentFilterValues := strings.Split(currFilter.input, "|")
				currFilter.values = currentFilterValues

				// finally append the filter or multiple to arr
				filtersArr = append(filtersArr, currFilter)
			}
		}
	}
	return filtersArr
}
