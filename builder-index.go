package finder

import (
	"context"
	"errors"
	"net/url"

	"github.com/Masterminds/squirrel"
)

var ErrNoSelectsSlice = errors.New("selects must be a pointer to a []string")

type ConfigIndex struct {
	// Timezone modifies the timezone of filters on the FromToCol timestamp
	// column values, the name is taken to be a location name corresponding to a
	// file in the IANA Time Zone database, such as "Africa/Tripoli".
	// defaults to 'UTC' when nil
	//
	// WHERE
	//     (table.created_at, table.created_at)
	// OVERLAPS (
	//  '2024-01-16T00:00:00Z'::TIMESTAMP AT TIME ZONE 'UTC',
	//  '2024-01-19T23:59:59Z'::TIMESTAMP AT TIME ZONE 'UTC'
	// )
	IDColumn  *string
	Timezone  *string
	FromToCol *string
	DB        Connection
	QB        *squirrel.StatementBuilderType
	// IsPublic is special where clause for "table.is_disabled" column
	// it can be used directly or add that to the wheres slice if the column
	// has another name for an active state
	IsPublic bool
	Joins    *[]string
	Selects  *[]string
	GroupBys *[]string
	Wheres   *[]squirrel.Sqlizer

	// PGInfo contains a map of table names
	// each key will have string slice containing sorted columns
	PGInfo map[string][]string
}

type configFinder struct {
	FromToCol *string
	GroupBys  *[]string
	IsPublic  bool
	Meta      *Meta
	Model     Model
	Results   *squirrel.SelectBuilder
	IDColumn  *string
	Timezone  *string
	UrlQuery  *URLQuery
	UrlValues *url.Values
}

func IndexBuilder[T Model](
	urlValues url.Values,
	c *ConfigIndex,
) (*IndexResponse[T], error) {
	var model T
	modelList := []T{}
	var indexResponse IndexResponse[T]

	tableName := model.TableName()
	if c.GroupBys == nil {
		c.GroupBys = &[]string{tableName + ".id"}
	}
	if c.Selects == nil {
		return nil, ErrNoSelectsSlice
	}
	results := c.QB.
		Select(*c.Selects...).
		From(tableName)
	if c.Joins != nil {
		for _, join := range *c.Joins {
			results = results.LeftJoin(join)
		}
	}
	if c.Wheres != nil {
		if len(*c.Wheres) != 0 {
			for _, where := range *c.Wheres {
				results = results.Where(where)
			}
		}
	}

	// parse filter, sort and q params ------ ---------------------------------
	var meta Meta
	meta.SearchColumns = model.SearchFields()
	meta.Columns = model.Columns(c.PGInfo)

	cf := &configFinder{
		GroupBys:  c.GroupBys,
		IsPublic:  c.IsPublic,
		Meta:      &meta,
		Model:     model,
		Results:   &results,
		IDColumn:  c.IDColumn,
		Timezone:  c.Timezone,
		UrlQuery:  &URLQuery{},
		UrlValues: &urlValues,
		FromToCol: c.FromToCol,
	}
	if err := find(cf); err != nil {
		return nil, err
	}

	// getting total count of filtered result ---------------------------------
	var total uint64
	totalQuery, args, err := c.QB.
		Select(`COUNT(*)`).
		FromSelect(results, "results").
		ToSql()
	if err != nil {
		return nil, err
	}
	row := c.DB.QueryRowContext(context.Background(), totalQuery, args...)
	if err := row.Scan(&total); err != nil {
		total = 0
	}
	meta.Total = total

	// choose to not paginate results -----------------------------------------
	showAllResults := urlValues.Has("all")
	if showAllResults {
		cf.UrlQuery.Page = 1
		cf.UrlQuery.Paginate = total
	}

	// build meta and query the results ---------------------------------------
	metaBuilder(&meta, cf.UrlQuery)

	query, args, err := results.
		Limit(meta.Paginate).
		Offset(meta.From - 1).
		ToSql()
	if err != nil {
		return nil, err
	}
	if err := c.DB.
		SelectContext(
			context.Background(),
			&modelList,
			query,
			args...,
		); err != nil {
		return &indexResponse, err
	}

	indexResponse.Meta = &meta
	indexResponse.Data = &modelList

	return &indexResponse, nil
}
