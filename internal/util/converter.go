package util

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
)

// ConvertToNullString converts a string to sql.NullString
func ConvertToNullString(input string) sql.NullString {
	return sql.NullString{
		String: input,
		Valid:  input != "",
	}
}

// ConvertToNullBool converts a bool to sql.NullBool
func ConvertToNullBool(input bool) sql.NullBool {
	return sql.NullBool{
		Bool:  input,
		Valid: true,
	}
}

// GenerateSlug generates a URL-friendly slug from a string
func GenerateSlug(input string) string {
	return slug.Make(input)
}

func parseStringToIntArray(input string) ([]int64, error) {
	if input == "" {
		return nil, nil
	}

	stringParts := strings.Split(input, ",")
	var intArray []int64

	for _, stringPart := range stringParts {
		trimmedString := strings.TrimSpace(stringPart)
		parsedInteger, err := strconv.ParseInt(trimmedString, 10, 64)
		if err != nil {
			return nil, err
		}
		intArray = append(intArray, parsedInteger)
	}

	return intArray, nil
}
