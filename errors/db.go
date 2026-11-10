package errors

import (
	"fmt"
)

type DB struct {
	fn  string
	err error
	pkg string
}

func (d DB) Error() string {
	return fmt.Sprintf("Database ERROR: %s - %s: %s", d.pkg, d.fn, d.err.Error())
}

func (d DB) Unwrap() error { return d.err }

func Database(packageName, function string, err error) error {
	return DB{
		fn:  function,
		pkg: packageName,
		err: err,
	}
}
