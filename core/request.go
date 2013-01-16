package din

import (
	"encoding/json"
	"fmt"
	"github.com/jordanorelli/din/dinutil"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var CSRFKey = "_din_csrf"

// the Request struct wraps the http.Request struct, providing a slice of
// strings representing the positional arguments found in a url pattern, and a
// map[string]string called kwargs representing the named parameters captured
// in url parsing.
type Request struct {
	// raw http request as accepted by the Din application server
	*http.Request

	// pointer to the route that satisfied the route filters for this http request
	*RouteMatch

	// each request is given an Id for log purposes.  The Id format is actually
	// the same as the mongodb Id format.
	Id RequestId

	// time that the request was received.
	Received time.Time

	logmux      sync.Mutex
	s           Session
	sessionKey  string
	newSession  bool
	saveSession bool
}

// parses an int from the query parameters found in the request.  The parameter
// required allows a developer to specify that the parameter is option.  If an
// optional parameter is not found in the query string, the min value is
// returned.
func (r *Request) BoundedInt(name string, required bool, min, max int) (int, error) {
	val_s := r.URL.Query().Get(name)

	if val_s == "" {
		if required {
			return 0, fmt.Errorf(`Missing required int parameter "%s"`, name)
		}
		return min, nil
	}

	val, err := strconv.Atoi(val_s)
	if err != nil {
		return 0, err
	}

	if val > max || val < min {
		return 0, fmt.Errorf("%s parameter out of range. Min: %d, Max: %d", name, min, max)
	}

	return val, nil
}

// Parses a time.Time from a unix epoch timestamp found in a named querystring.
// Technically a querystring key can appear multiple times; in that case, only
// the first value is observed.  The bool "required" allows us to mark the
// parameter as optional (really, I should rename this to optional and flip it
// where I use it).  If the timestamp parameter is optional and not supplied,
// the zero value for time.Time is returned.
func (r *Request) Timestamp(name string, required bool) (time.Time, error) {
	val_s := r.URL.Query().Get(name)

	if val_s == "" {
		if required {
			return *new(time.Time), fmt.Errorf(`Missing required timestamp parameter "%s"`, name)
		}
		return *new(time.Time), nil
	}

	val, err := strconv.ParseInt(val_s, 10, 0)
	if err != nil {
		return *new(time.Time), err
	}

	return time.Unix(val, 0), nil
}

func (r *Request) UsingSSL() bool {
	header := r.Header.Get("X-Forwarded-Ssl")
	ret := (header != "" && header == "on")
	return ret
}

func (r *Request) Abspath(relpath string) string {
	if r.UsingSSL() {
		return "https://" + path.Join(r.Host, relpath)
	}
	return "http://" + path.Join(r.Host, relpath)
}

func (r *Request) Addpath(relpath string) string {
	if r.UsingSSL() {
		return "https://" + path.Join(r.Host, r.URL.Path, relpath)
	}
	return "http://" + path.Join(r.Host, r.URL.Path, relpath)
}

// NextUrl constructs new urls based on the url found in the existing http
// request.  Effectively, you can use it to substitute values in a request's
// querystring to generate pagination urls.
func (r *Request) NextUrl(params map[string]string) string {
	uri := r.URL
	q := uri.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	uri.RawQuery = q.Encode()
	return r.Abspath(uri.String())
}

func (r *Request) LogReceived() (int, error) {
	r.logmux.Lock()
	defer r.logmux.Unlock()
	return fmt.Println("-->", time.Now().Unix(), r.Id, r.Method, r.URL)
}

func (r *Request) LogError(err error) {
	var statusCode int
	switch e := err.(type) {
	case Error:
		statusCode = e.StatusCode
	default:
		statusCode = http.StatusInternalServerError
	}
	fmt.Println(statusCode, time.Now().Unix(), r.Id, statusCode, time.Since(r.Received), err.Error())
}

func (r *Request) LogTimeout() {
	fmt.Println(http.StatusGatewayTimeout, time.Now().Unix(), r.Id, time.Since(r.Received))
}

func (r *Request) LogResponse(statusCode int) {
	fmt.Println(statusCode, time.Now().Unix(), r.Id, time.Since(r.Received))
}

func (r *Request) LogPanic(v interface{}) {
	fmt.Println(500, time.Now().Unix(), r.Id, 500, time.Since(r.Received), "PANIC", v)
}

func (r *Request) Log(msg string) {
	fmt.Println("...", time.Now().Unix(), r.Id, msg)
}

func (r *Request) Logf(format string, v ...interface{}) {
	format = fmt.Sprintf("... %v %v %v\n", time.Now().Unix(), r.Id, format)
	fmt.Printf(format, v...)
}

func (r *Request) UnmarshalJSON(v interface{}) error {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		return Error{
			StatusCode: http.StatusBadRequest,
			Message:    `invalid json input`,
		}
	}
	return nil
}

func (r *Request) SessionGet(key string, dest interface{}) error {
	s, err := r.session()
	if err != nil {
		return err
	}
	return s.get(key, dest)
}

func (r *Request) SessionClear() {
	key, err := r.SessionKey()
	if err == nil {
		sessions.Delete(key)
	}
}

func (r *Request) createSession() {
	r.s = make(Session, 5)
	r.newSession = true
	r.sessionKey = dinutil.RandomString(32)
}

func (r *Request) SessionSet(key string, v interface{}) {
	if r.s == nil {
		r.createSession()
	}
	r.s.set(key, v)
	r.saveSession = true
}

func (r *Request) session() (Session, error) {
	if r.s != nil {
		return r.s, nil
	}
	key, err := r.SessionKey()
	if err != nil {
		return nil, err
	}
	return sessions.Get(key)
}

func (r *Request) SessionKey() (string, error) {
	if r.sessionKey != "" {
		return r.sessionKey, nil
	}
	cookie, err := r.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return "", ErrNoSessionId
	}
	/*
		if !cookie.HttpOnly {
			r.Log("WARN: potential false cookie attempt: cookie is not http only")
			return "", ErrInvalidSessionCookie
		}
		if !cookie.Secure {
			r.Log("WARN: potential false cookie attempt: cookie is not secure")
			return "", ErrInvalidSessionCookie
		}
	*/
	r.sessionKey = cookie.Value
	return cookie.Value, nil
}

// A RequestItem is defined as any item that may define how it can be read from
// an incoming http request.
type RequestItem interface {
	FromRequest(*Request) error
}

// Given a RequestItem, unpack the item from the request.  It basically just
// calls the RequestItem's FromRequest method; it's just a convenience item.
// In the future it may take an interface{} as input and attempt to unpack the
// item automatically if the item does not implement its own unpacking
// strategy.
func (r *Request) Unpack(i RequestItem) error {
	return i.FromRequest(r)
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return `din: Unmarshal(nil)`
	}
	if e.Type.Kind() != reflect.Ptr {
		return `din: Unmarshal(non-pointer ` + e.Type.String() + `)`
	}
	return `din: Unmarshal(nil ` + e.Type.String() + `)`
}

type Unmarshaler interface {
	Unmarshal(string) error
}

func (r *Request) QueryVal(key string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.IsNil() || rv.Kind() != reflect.Ptr {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	if rv.Type().NumMethod() > 0 {
		if u, isUnmarshaler := rv.Interface().(Unmarshaler); isUnmarshaler {
			s := r.URL.Query().Get(key)
			if s == "" {
				return fmt.Errorf(`din: key missing in Request.QueryVal`)
			}
			return u.Unmarshal(s)
		}
	}

	switch v.(type) {
	case *int:
		s := r.URL.Query().Get(key)
		if s == "" {
			return fmt.Errorf(`din: key missing in Request.QueryVal`)
		}

		i, err := strconv.ParseInt(s, 10, 0)
		if err != nil {
			return err
		}
		reflect.Indirect(rv).SetInt(i)
		return nil

	case *string:
		s := r.URL.Query().Get(key)
		if s == "" {
			return fmt.Errorf(`din: key missing in Request.QueryVal`)
		}

		reflect.Indirect(rv).SetString(s)
		return nil

	case *bool:
		s, ok := r.URL.Query()[key]
		if !ok {
			return nil
		}
		if len(s) == 0 {
			return nil
		}
		switch strings.ToLower(s[0]) {
		case "false", "f", "no", "nay", "non":
			return nil
		}
		reflect.Indirect(rv).SetBool(true)
	}
	return nil
}

func KwargMissing(key string) Error {
	return Error{
		StatusCode: http.StatusBadRequest,
		Message:    fmt.Sprintf("missing required kwarg %v", key),
	}
}

func ParseApiVersion(s string) (ClientApiVersion, error) {
	t, err := time.Parse("20060102", s)
	return ClientApiVersion(t), err
}

// Parses the requested ApiVersion from the query string. If no value is
// present or the value cannot be parsed the default value ApiOldestVersion is
// returned.
// Query value read is 'v', format is 'YYYYMMDD'
// Inspired by https://developer.foursquare.com/overview/versioning
func (r *Request) ApiVersion() ClientApiVersion {
	var s string
	if err := r.QueryVal("v", &s); err != nil {
		return ApiOldestVersion
	}

	version, err := ParseApiVersion(s)
	if err != nil {
		return ApiOldestVersion
	}

	return ClientApiVersion(version)
}

// tells us whether a request is signed with an csrf token or not.  Right now,
// this doesn't actually do any *security*, it just checks for the existence of
// a _din_csrf cookie.
func (r *Request) IsSigned() bool {
	if _, err := r.Cookie(CSRFKey); err == nil {
		return true
	}
	if v := r.URL.Query().Get(CSRFKey); v != "" {
		return true
	}
	return false
}

type ClientApiVersion time.Time

func (v ClientApiVersion) String() string {
	t := time.Time(v)
	return t.Format("20060102")
}

var ApiOldestVersion ClientApiVersion
