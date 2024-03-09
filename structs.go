package finder

import (
	"errors"
	"fmt"
)

const ZeroedUUID = "00000000-0000-0000-0000-000000000000"

var (
	ErrRequiresTransaction    = errors.New("requires a transaction")
	ErrNoProvidedID           = errors.New("no provided id")
	ErrSelectNotAPointerSlice = errors.New(
		"selects must be a pointer to []string",
	)
	ErrInsertsNotAPointerSlice = errors.New(
		"inserts must be a pointer to []string",
	)
	ErrDefaultColsNotAPointerSlice = errors.New(
		"defaultCols must be a pointer to []string",
	)
	ErrInputNotAPointerSlice = errors.New(
		"inserts must be a pointer to []any",
	)
	ErrWhereMustHaveElementsWhenNoID = errors.New(
		"selecting an element without id must have at least one where clause",
	)
)

func ErrInputLengthMismatch(input *[]any, defaultCols *[]string) error {
	if input == nil {
		return ErrInputNotAPointerSlice
	}
	if defaultCols == nil {
		return ErrDefaultColsNotAPointerSlice
	}
	return fmt.Errorf(
		"input [%d] column count does not match args [%d]",
		len(*input),
		len(*defaultCols),
	)
}

type Meta struct {
	Total         uint64    `json:"total"`
	Paginate      uint64    `json:"per_page"`
	CurrentPage   uint64    `json:"current_page"`
	FirstPage     uint64    `json:"first_page"`
	LastPage      uint64    `json:"last_page"`
	From          uint64    `json:"from"`
	To            uint64    `json:"to"`
	Columns       *[]string `json:"columns"`
	SearchColumns *[]string `json:"search_columns"`
}

type IndexResponse[T any] struct {
	Meta *Meta `json:"meta"`
	Data *[]T  `json:"data"`
}

type URLQuery struct {
	Q        string `query:"q"`
	Page     uint64 `query:"page"`
	Sorts    string `query:"sorts"`
	Filters  string `query:"filters"`
	Paginate uint64 `query:"paginate"`
}

type Join struct {
	From, To string
}

type Through struct {
	Table string
	Join  *Join
}

type RelationField struct {
	Table   string
	Join    *Join
	Through *Through
}
