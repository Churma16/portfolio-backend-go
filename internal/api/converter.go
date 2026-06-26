package api

import "github.com/jackc/pgx/v5/pgtype"

// Helper mengubah string biasa jadi pgtype.Text
func convertToNullString(val string) pgtype.Text {
	return pgtype.Text{
		String: val,
		Valid:  val != "",
	}
}

// Helper mengubah bool biasa jadi pgtype.Bool
func convertToNullBool(val bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  val,
		Valid: true, // Kita anggap selalu valid kalau dikirim true/false
	}
}

// Helper mengubah int64 biasa jadi pgtype.Int8
func convertToNullInt64(val int64) pgtype.Int8 {
	return pgtype.Int8{
		Int64: val,
		Valid: val != 0,
	}
}
