package finder

import (
	"context"

	"github.com/Masterminds/squirrel"
)

type ConfigStore struct {
	DB         Connection
	QB         *squirrel.StatementBuilderType
	TableName  string
	TableAlias string
	Input      *[]any
	Inserts    *[]string
	Selects    *[]string
	Joins      *[]string
}

func CreateOne(created Model, c *ConfigStore) error {
	if c.TableName == "" {
		c.TableName = created.TableName()
	}
	if c.TableAlias == "" {
		c.TableAlias = created.TableName()
	}
	if c.Inserts == nil {
		return ErrInsertsNotAPointerSlice
	}
	if c.Input == nil {
		return ErrInputNotAPointerSlice
	}
	subquery := c.QB.
		Insert(c.TableName).
		Suffix("RETURNING *").
		Columns(*c.Inserts...).
		Values(*c.Input...)
	with := subquery.
		Prefix("WITH " + c.TableAlias + " AS (").
		Suffix(")")

	result := c.QB.Select(*c.Selects...).
		PrefixExpr(with).
		From(c.TableAlias)
	if c.Joins != nil {
		if len(*c.Joins) > 0 {
			for _, join := range *c.Joins {
				result = result.LeftJoin(join)
			}
		}
	}
	query, args, err := result.ToSql()
	if err != nil {
		return err
	}
	return c.DB.GetContext(
		context.Background(),
		created,
		query,
		args...,
	)
}
