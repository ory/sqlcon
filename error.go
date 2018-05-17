package sqlcon

import (
	"net/http"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/ory/herodot"
	"github.com/pkg/errors"
)

var (
	ErrUniqueViolation = &herodot.DefaultError{
		CodeField:   http.StatusBadRequest,
		StatusField: http.StatusText(http.StatusBadRequest),
		ErrorField:  "Unable to insert row because a column value is not unique",
	}
)

func HandleError(err error) error {
	if err, ok := err.(*pq.Error); ok {
		switch err.Code.Name() {
		case "unique_violation":
			return errors.Wrap(ErrUniqueViolation, err.Error())
		}
		return errors.WithStack(err)
	}

	if err, ok := err.(*mysql.MySQLError); ok {
		return errors.WithStack(err)
	}

	return errors.WithStack(err)
}
