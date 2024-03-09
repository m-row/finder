package finder

import (
	"context"

	"github.com/Masterminds/squirrel"
)

type OptimisticLock struct {
	Name  string
	Value any
}

type ConfigUpdate struct {
	DB             Connection
	QB             *squirrel.StatementBuilderType
	TableName      string
	TableAlias     string
	OptimisticLock *OptimisticLock
	Input          *[]any
	Inserts        *[]string
	Selects        *[]string
	Joins          *[]string
	Wheres         *[]squirrel.Sqlizer
}

func UpdateOne(updated Model, c *ConfigUpdate) error {
	if c.TableName == "" {
		c.TableName = updated.TableName()
	}
	if c.TableAlias == "" {
		c.TableAlias = updated.TableName()
	}
	if c.OptimisticLock != nil {
		if c.OptimisticLock.Name == "" {
			c.OptimisticLock.Name = "updated_at"
		}
	}
	if c.Inserts == nil {
		return ErrInsertsNotAPointerSlice
	}
	if c.Input == nil {
		return ErrInputNotAPointerSlice
	}
	id := updated.GetID()
	if id == "" || id == "0" || id == ZeroedUUID {
		return ErrNoProvidedID
	}

	subquery := c.QB.
		Update(c.TableName).
		Suffix("RETURNING *").
		Where("id=?", id)
	if c.OptimisticLock != nil {
		subquery = subquery.Where(
			c.OptimisticLock.Name+" = ?",
			c.OptimisticLock.Value,
		)
	}

	if c.Wheres != nil {
		if len(*c.Wheres) != 0 {
			for _, where := range *c.Wheres {
				subquery = subquery.Where(where)
			}
		}
	}
	for i, column := range *c.Inserts {
		subquery = subquery.Set(column, (*c.Input)[i])
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
		updated,
		query,
		args...,
	)
}
