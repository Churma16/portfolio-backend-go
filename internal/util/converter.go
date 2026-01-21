package util

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

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

func ParseInterfaceToIntArray(data interface{}) []int64 {
	var result []int64

	switch v := data.(type) {
	case []interface{}:
		// If it's already an array
		for _, item := range v {
			if floatVal, ok := item.(float64); ok {
				result = append(result, int64(floatVal))
			}
		}
	case string:
		// If it's a string, parse it
		if parsedArray, err := ParseStringToIntArray(v); err == nil {
			for _, val := range parsedArray {
				result = append(result, int64(val))
			}
		}
	}

	return result
}

func ParseStringToIntArray(input string) ([]int64, error) {
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

func FormatDate(nullString sql.NullString, layout string) string {
	if !nullString.Valid {
		return ""
	}

	possibleLayouts := []string{
		"2006-01-02",
		"2006-01",
		"2006",
	}

	for _, possibleLayout := range possibleLayouts {
		parsedDate, err := time.Parse(possibleLayout, nullString.String)
		if err == nil {
			return parsedDate.Format(layout)
		}
	}

	return ""
}
