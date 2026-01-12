package api

import "database/sql"

// Helper mengubah string biasa jadi sql.NullString
func convertToNullString(val string) sql.NullString {
	return sql.NullString{
		String: val,
		Valid:  val != "",
	}
}

// Helper mengubah bool biasa jadi sql.NullBool
func convertToNullBool(val bool) sql.NullBool {
	return sql.NullBool{
		Bool:  val,
		Valid: true, // Kita anggap selalu valid kalau dikirim true/false
	}
}
