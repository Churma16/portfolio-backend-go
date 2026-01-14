package util

import (
	"database/sql"
	"strconv"
	"strings"

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

func parseStringToIntArray(input string) ([]int64, error) {
	if input == "" {
		return nil, nil
	}

	stringParts := strings.Split(input, ",")
	var intArray []int64

	for _, part := range stringParts {
		trimmedPart := strings.TrimSpace(part)
		parsedInt, err := strconv.ParseInt(trimmedPart, 10, 64)
		if err != nil {
			return nil, err
		}
		intArray = append(intArray, parsedInt)
	}

	return intArray, nil
}
