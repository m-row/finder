package finder

import (
	"net/url"
)

// Model interface reflective of database objects.
type Model interface {
	// GetID returns id of model as a string
	GetID() string
	// ModelName returns text name of model i18n-able
	ModelName() string
	// TableName returns text name of table of model NOT i18n-able
	TableName() string
	// DefaultSearch returns column name used with gofinder
	DefaultSearch() string
	// SearchFields returns columns slice of searchable fields used with
	// gofinder can be singular columns of model:
	//  "name", or related as: "rel:customers.name",
	SearchFields() *[]string
	// Relations returns table, value array of relations used with gofinder
	Relations() *[]RelationField
	// Columns should return a sorted list of table columns, verifies
	// values for sorting and filtering, ideally a query on app startup
	// loads the info and then they are mapped to the model by this func
	Columns(map[string][]string) *[]string
	// Initialize handles model id input or generation if uuid and sequence
	// from conn if it is an int, used heavily for testing
	// input of id is not required and the db handles default valuing
	Initialize(url.Values, Connection) bool
}

func GetColumns(m Model, pgInfo map[string][]string) *[]string {
	cols, found := pgInfo[m.TableName()]
	if !found {
		return &[]string{}
	}
	return &cols
}
