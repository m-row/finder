package finder

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type ConfigDelete struct {
	DB         Connection
	QB         *squirrel.StatementBuilderType
	TableName  string
	TableAlias string
	IDColumn   *string
	Selects    *[]string
	Joins      *[]string
	Wheres     *[]squirrel.Sqlizer
}

func DeleteOne(deleted Model, c *ConfigDelete) error {
	if c.TableName == "" {
		c.TableName = deleted.TableName()
	}
	if c.TableAlias == "" {
		c.TableAlias = deleted.TableName()
	}
	id := deleted.GetID()
	if id == "" || id == "0" || id == ZeroedUUID {
		return ErrNoProvidedID
	}
	if c.Selects == nil {
		return ErrSelectNotAPointerSlice
	}

	subquery := c.QB.
		Delete(c.TableName).
		Suffix("RETURNING *")
	if c.IDColumn != nil {
		w := fmt.Sprintf("%s=?", *c.IDColumn)
		subquery = subquery.Where(w, id)
	} else {
		subquery = subquery.Where("id=?", id)
	}
	if c.Wheres != nil {
		if len(*c.Wheres) != 0 {
			for _, where := range *c.Wheres {
				subquery = subquery.Where(where)
			}
		}
	}
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
		deleted,
		query,
		args...,
	)
}
