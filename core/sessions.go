package din

import (
	"net/http"
	"reflect"
	// "time"
)

const SESSION_COOKIE_NAME = "_din_session"

var sessions SessionHandler

type Session map[string]interface{}

func (s Session) get(key string, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.IsNil() || destValue.Kind() != reflect.Ptr {
		return &InvalidUnmarshalError{reflect.TypeOf(dest)}
	}
	storedVal, ok := s[key]
	if !ok {
		return ErrInvalidSessionKey
	}
	if !reflect.Indirect(destValue).CanSet() {
		return ErrInvalidSessionDest
	}
	if reflect.Indirect(destValue).Type() != reflect.ValueOf(storedVal).Type() {
		return ErrInvalidSessionDest
	}
	reflect.Indirect(destValue).Set(reflect.ValueOf(storedVal))
	return nil
}

func (s Session) set(key string, val interface{}) {
	s[key] = val
}

type SessionHandler interface {
	Get(string) (Session, error)
	Set(string, Session) error
	Delete(string) error
}

func SetSessionHandler(handler SessionHandler) {
	sessions = handler
}

type defaultSessionHandler map[string]Session

func (d defaultSessionHandler) Get(id string) (Session, error) {
	s, ok := d[id]
	if !ok {
		return nil, ErrUnknownSessionId
	}
	return s, nil
}

func (d defaultSessionHandler) Set(id string, s Session) error {
	d[id] = s
	return nil
}

func (d defaultSessionHandler) Delete(id string) error {
	delete(d, id)
	return nil
}

func init() {
	if sessions == nil {
		SetSessionHandler(make(defaultSessionHandler))
	}
}

func setSessionId(w http.ResponseWriter, id string) {
	http.SetCookie(w, &http.Cookie{
		Name:  SESSION_COOKIE_NAME,
		Path:  "/",
		Value: id,
		// Domain:   "",
		// HttpOnly: true,
		// Secure:   true,
	})
}
