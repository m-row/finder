package finder

import (
	"context"

	"github.com/Masterminds/squirrel"
)

type ConfigShow struct {
	DB         Connection
	QB         *squirrel.StatementBuilderType
	TableName  string
	TableAlias string
	// IsPublic specifically used to check where table_name.is_disabled field
	// is toggled off
	IsPublic bool
	Selects  *[]string
	Joins    *[]string
	Wheres   *[]squirrel.Sqlizer
}

func ShowOne(shown Model, c *ConfigShow) error {
	if c.TableName == "" {
		c.TableName = shown.TableName()
	}
	if c.TableAlias == "" {
		c.TableAlias = shown.TableName()
	}
	id := shown.GetID()
	noID := id == "" || id == "0" || id == ZeroedUUID
	if c.Wheres == nil && noID {
		return ErrWhereMustHaveElementsWhenNoID
	}
	if noID && len(*c.Wheres) == 0 {
		return ErrNoProvidedID
	}
	if c.Selects == nil {
		return ErrSelectNotAPointerSlice
	}
	result := c.QB.
		Select(*c.Selects...).
		From(c.TableName)

	if !noID {
		result = result.Where(c.TableAlias+".id=?", id)
	}

	if c.Wheres != nil {
		if len(*c.Wheres) != 0 {
			for _, where := range *c.Wheres {
				result = result.Where(where)
			}
		}
	}
	if c.IsPublic {
		result = result.Where(c.TableAlias + ".is_disabled=false")
	}
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
	return c.DB.GetContext(context.Background(), shown, query, args...)
}
