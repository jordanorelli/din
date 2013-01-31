package din

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNoSessionId          = errors.New("request has no session id")
	ErrInvalidSessionKey    = errors.New("invalid session key")
	ErrInvalidSessionDest   = errors.New("invalid session value destination")
	ErrInvalidSessionCookie = errors.New("invalid session cookie")
	ErrUnknownSessionId     = errors.New("unknown session id")
	ErrQueryKeyMissing      = Error{
		StatusCode: http.StatusBadRequest,
		Message:    `din: querykey missing`,
	}
)

// the din.Error type is to be used for errors that can be rendered to be shown
// to the user.  If a din.Error is the last error in a pipeline, the error's
// StatusCode will be used to inform the proper HTTP status code.  The message
// is assumed to be safe to be shown to a user.  (future versions will likely
// include another unsafe message field for logging, but I haven't gotten there
// yet.)
type Error struct {
	StatusCode int
	Message    string
}

func (e Error) Error() string { return e.Message }

func InternalServerError(format string, vals ...interface{}) error {
	return Error{
		StatusCode: http.StatusInternalServerError,
		Message:    fmt.Sprintf(format, vals...),
	}
}

func StatusForbidden(format string, vals ...interface{}) error {
	return Error{
		StatusCode: http.StatusForbidden,
		Message:    fmt.Sprintf(format, vals...),
	}
}

func ErrUnknownHandler(name string) error {
	return errors.New("unknown handler " + name)
}

func ErrUnkownTemplate(relpath string) error {
	return errors.New("unknown template " + relpath)
}

func ErrorStage(err error) Stage {
	return func(*Request) (Response, error) {
		return nil, err
	}
}
