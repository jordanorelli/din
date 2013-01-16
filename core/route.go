package din

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sync/atomic"
	"time"
)

// PanicDepth defines the stack depth that should be expected when rendering
// stack traces inside of a panic handler.
const PanicDepth = 3

type PanicHandler func(http.ResponseWriter, *Request, chan struct{})

type ErrorHandler func(http.ResponseWriter, *Request, error)

type NotFoundHandler func(http.ResponseWriter, *Request)

// struct Router implements http.Handler, so that it may be used with the
// default http library.  It keeps a registry mapping regexes to functions for
// easier url parsing.
type Router struct {
	OnPanic         PanicHandler
	OnError         ErrorHandler
	On404           NotFoundHandler
	routes          []*Pipeline
	staticWhitelist []string
	staticPaths     []staticPath
	started         time.Time
}

// struct Pipeline defines a series of handlers to be registered for a given
// URL pattern.  The http request is passed to each subsequent "stage" in the
// pipeline.
type Pipeline struct {
	Route    *Route  `json:"route"`
	Name     string  `json:"name"`
	Doc      string  `json:"doc"`
	Handlers []Stage `json:"handlers"`
}

func (p *Pipeline) String() string {
	return fmt.Sprintf("{Pipeline Name: %s, Handlers: %v}", p.Name, p.Handlers)
}

func NewRouter(errorHandler ErrorHandler, panicHandler PanicHandler) *Router {
	if errorHandler == nil {
		errorHandler = DefaultErrorHandler
	}
	if panicHandler == nil {
		panicHandler = DefaultPanicHandler
	}
	return &Router{
		OnPanic: panicHandler,
		OnError: errorHandler,
		routes:  []*Pipeline{},
		started: time.Now(),
	}
}

func DefaultPanicHandler(w http.ResponseWriter, r *Request, p chan struct{}) {
	if recovered := recover(); recovered != nil {
		defer close(p)
		fmt.Fprint(w, recovered)
	}
}

func JSONPanicHandler(w http.ResponseWriter, r *http.Request) {
	if recovered := recover(); recovered != nil {
		s := make([]uintptr, 10)
		n := runtime.Callers(PanicDepth, s)
		type trace struct {
			Name string `json:"name"`
			File string `json:"file"`
			Line int    `json:"line"`
		}
		deets := make([]trace, n)
		for i := 0; i < n; i++ {
			f := runtime.FuncForPC(s[i])
			deets[i].Name = f.Name()
			deets[i].File, deets[i].Line = f.FileLine(s[i])
		}
		raw, err := json.MarshalIndent(struct {
			Recovered interface{} `json:"recovered"`
			Trace     []trace     `json:"trace"`
		}{recovered, deets}, "", "  ")
		if err != nil {
			io.WriteString(w, "whyyyy")
			return
		}
		w.Write(raw)
	}
}

func DefaultErrorHandler(w http.ResponseWriter, req *Request, err error) {
	switch e := err.(type) {
	case Error:
		w.WriteHeader(e.StatusCode)
		io.WriteString(w, e.Message)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, e.Error())
	}
}

// implements the http.Handler interface, so that we may use our router with
// the default http package.
func (r *Router) ServeHTTP(w http.ResponseWriter, raw *http.Request) {
	c, errchan, p := make(chan Response), make(chan error), make(chan struct{})
	req := r.match(raw)

	if req.RouteMatch == nil {
		req.LogResponse(http.StatusNotFound)
		if r.On404 != nil {
			r.On404(w, req)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404"))
		return
	}

	go func() {
		defer r.OnPanic(w, req, p)
		for _, fn := range req.Pipeline.Handlers {
			res, err := fn(req)
			if err != nil {
				errchan <- err
				return
			}
			if res != nil {
				c <- res
				return
			}
		}
	}()

	req.Logf("route: %v", req.RouteMatch.Pipeline.Name)

	select {
	case <-time.After(30 * time.Second):
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("herp derp, i timed out"))
		req.LogTimeout()
	case res := <-c:
		if req.newSession {
			setSessionId(w, req.sessionKey)
		}
		if err := res.Render(w); err != nil {
			req.LogError(err)
			break
		}
		if req.saveSession {
			if err := sessions.Set(req.sessionKey, req.s); err != nil {
				req.LogError(err)
			}
		}
		req.LogResponse(res.Status())
	case err := <-errchan:
		r.OnError(w, req, err)
		req.LogError(err)
	case <-p:
		break
	}
}

func (r *Router) ListenAndServe(addr string) error {
	server := &http.Server{Addr: addr, Handler: r}
	return server.ListenAndServe()
}

// transforms an incoming http.Request into a din.Request.  If a route match is
// found for this request, it is stored in the request; otherwise it is nil.
func (r *Router) match(raw *http.Request) *Request {
	req := &Request{
		Request:  raw,
		Id:       newRequestId(),
		Received: time.Now(),
	}

	req.LogReceived() // TODO: observe returned error val
	for _, e := range r.routes {
		if match := e.Route.Match(raw.URL.Path); match != nil {
			req.RouteMatch = match
			req.Pipeline = e
			return req
		}
	}
	return req
}

type Stage func(*Request) (Response, error)

func (s *Stage) UnmarshalJSON(b []byte) error {
	var name string
	if err := json.Unmarshal(b, &name); err != nil {
		return err
	}
	found, ok := getHandler(name)
	if !ok {
		return ErrUnknownHandler(name)
	}
	*s = found
	return nil
}

// a Response is any struct that can be Rendered into an http response.
type Response interface {
	Render(http.ResponseWriter) error
	Status() int
}

func (router *Router) AddRoute(pattern string, name string, stages ...Stage) {
	router.routes = append(router.routes, &Pipeline{
		Route:    NewRoute(pattern),
		Name:     name,
		Handlers: stages,
	})
}

type RouteMatch struct {
	// positional arguments found in the route's regex pattern
	Args []string
	// keyword arguments taken from named capture groups found in the route's regex pattern
	Kwargs map[string]string
	*Pipeline
}

// right now, just embeds a regex.  A "name" field should also be added here.
type Route struct {
	*regexp.Regexp
}

func NewRoute(pattern string) *Route {
	return &Route{regexp.MustCompile(pattern)}
}

func (r *Route) Match(target string) *RouteMatch {
	submatches := r.FindStringSubmatch(target)
	if submatches == nil {
		return nil
	}

	if len(submatches) == 1 {
		return new(RouteMatch)
	}

	m := new(RouteMatch)
	submatches = submatches[1:]
	for i, name := range r.SubexpNames()[1:] {
		if name == "" {
			m.Args = append(m.Args, submatches[i])
		} else {
			if m.Kwargs == nil {
				m.Kwargs = make(map[string]string)
			}
			m.Kwargs[name] = submatches[i]
		}
	}
	return m
}

func (r *Route) UnmarshalJSON(b []byte) error {
	var pattern string
	if err := json.Unmarshal(b, &pattern); err != nil {
		return err
	}
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	*r = Route{regex}
	return nil
}

// RequestId is used for tagging each incoming http request for logging
// purposes.  The actual implementation is just the ObjectId implementation
// found in launchpad.net/mgo/bson.  This will most likely change and evolve
// into its own format.
type RequestId string

func (id RequestId) String() string {
	return fmt.Sprintf("%x", string(id))
}

// Time returns the timestamp part of the id.
// It's a runtime error to call this method with an invalid id.
func (id RequestId) Time() time.Time {
	secs := int64(binary.BigEndian.Uint32(id.byteSlice(0, 4)))
	return time.Unix(secs, 0)
}

// byteSlice returns byte slice of id from start to end.
// Calling this function with an invalid id will cause a runtime panic.
func (id RequestId) byteSlice(start, end int) []byte {
	if len(id) != 12 {
		panic(fmt.Sprintf("Invalid RequestId: %q", string(id)))
	}
	return []byte(string(id)[start:end])
}

// requestIdCounter is atomically incremented when generating a new ObjectId
// using NewObjectId() function. It's used as a counter part of an id.
var requestIdCounter uint32 = 0

// machineId stores machine id generated once and used in subsequent calls
// to NewObjectId function.
var machineId []byte

// initMachineId generates machine id and puts it into the machineId global
// variable. If this function fails to get the hostname, it will cause
// a runtime error.
func initMachineId() {
	var sum [3]byte
	hostname, err := os.Hostname()
	if err != nil {
		panic("Failed to get hostname: " + err.Error())
	}
	hw := md5.New()
	hw.Write([]byte(hostname))
	copy(sum[:3], hw.Sum(nil))
	machineId = sum[:]
}

// NewObjectId returns a new unique ObjectId.
// This function causes a runtime error if it fails to get the hostname
// of the current machine.
func newRequestId() RequestId {
	b := make([]byte, 12)
	// Timestamp, 4 bytes, big endian
	binary.BigEndian.PutUint32(b, uint32(time.Now().Unix()))
	// Machine, first 3 bytes of md5(hostname)
	if machineId == nil {
		initMachineId()
	}
	b[4] = machineId[0]
	b[5] = machineId[1]
	b[6] = machineId[2]
	// Pid, 2 bytes, specs don't specify endianness, but we use big endian.
	pid := os.Getpid()
	b[7] = byte(pid >> 8)
	b[8] = byte(pid)
	// Increment, 3 bytes, big endian
	i := atomic.AddUint32(&requestIdCounter, 1)
	b[9] = byte(i >> 16)
	b[10] = byte(i >> 8)
	b[11] = byte(i)
	return RequestId(b)
}

var ErrBadMethod = Error{
	StatusCode: http.StatusMethodNotAllowed,
	Message:    "unsupported http method",
}

type VerbMux map[string]Stage

func (m VerbMux) Stage() Stage {
	return func(req *Request) (Response, error) {
		if stage, ok := m[req.Method]; ok {
			return stage(req)
		}
		return nil, ErrBadMethod
	}
}

// ParseRouteFile takes a path to a routes configuration file and returns a
// valid Router if it is able.  Otherwise, an error is returned.  Right now,
// the only supported format is json, which is a terrible format.
func ParseRouteFile(path string) (*Router, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	routes := make([]*Pipeline, 0)
	if err := json.NewDecoder(fi).Decode(&routes); err != nil {
		return nil, err
	}
	router := NewRouter(nil, nil)
	router.routes = routes
	return router, nil
}
