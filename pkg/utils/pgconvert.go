package utils

import "github.com/jackc/pgx/v5/pgtype"

type _pgConvert struct{}

// Create a global instance
var PgConvert = _pgConvert{}

func (_pgConvert) ToPgText(str string) (ret pgtype.Text) {
	if str != "" {
		ret.String = str
		ret.Valid = true
		return
	}
	return

}
