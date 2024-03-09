package finder

import (
	"context"
	"database/sql"
)

// Connection is an interface that executes sql queries, it is based on *sqlx.DB
type Connection interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	GetContext(context.Context, interface{}, string, ...interface{}) error
	Select(interface{}, string, ...interface{}) error
	SelectContext(context.Context, interface{}, string, ...interface{}) error
	QueryRowContext(context.Context, string, ...any) *sql.Row
	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}
