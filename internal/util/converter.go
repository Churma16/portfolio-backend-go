package util

import (
	"database/sql"

	"github.com/gosimple/slug"
)

// ConvertToNullString converts a string to sql.NullString
func ConvertToNullString(val string) sql.NullString {
	return sql.NullString{
		String: val,
		Valid:  val != "",
	}
}

// ConvertToNullBool converts a bool to sql.NullBool
func ConvertToNullBool(val bool) sql.NullBool {
	return sql.NullBool{
		Bool:  val,
		Valid: true,
	}
}

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(name string) string {
	return slug.Make(name)
}
