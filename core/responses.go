package din

import (
	"io"
	"net/http"
	"strings"
)

type emptyResponse struct {
	statusCode int
}

func (e emptyResponse) Render(w http.ResponseWriter) error {
	w.WriteHeader(e.statusCode)
	return nil
}

func (e emptyResponse) Status() int {
	return e.statusCode
}

func EmptyResponse(statusCode int) Response {
	return emptyResponse{statusCode}
}

type proxyResponse struct{ *http.Response }

func (res *proxyResponse) Render(w http.ResponseWriter) error {
	if res.Response.Header != nil {
		for k, v := range res.Response.Header {
			w.Header()[k] = v
		}
	}
	w.WriteHeader(res.Response.StatusCode)
	_, err := io.Copy(w, res.Response.Body)
	return err
}

func (res *proxyResponse) Status() int {
	return res.StatusCode
}

func ProxyResponse(res *http.Response) Response {
	return &proxyResponse{res}
}

// PlaintextResponse represents a ... plaintext response, in HTTP format.  This
// will output the contents of the io.Reader to an http.ResponseWriter, setting
// the content-type to text/plain, and using the desired StatusCode.
type PlaintextResponse struct {
	io.Reader
	StatusCode int
}

func (res *PlaintextResponse) Render(w http.ResponseWriter) error {
	h := w.Header()
	h.Set("Content-Type", "text/plain")
	if res.StatusCode != 0 {
		w.WriteHeader(res.StatusCode)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, err := io.Copy(w, res)
	return err
}

func (res *PlaintextResponse) Status() int {
	return res.StatusCode
}

func PlaintextResponseString(s string, code int) Response {
	return &PlaintextResponse{strings.NewReader(s), code}
}

type redirect struct {
	request *http.Request
	url     string
	code    int
}

func (r *redirect) Render(w http.ResponseWriter) error {
	http.Redirect(w, r.request, r.url, r.code)
	return nil
}

func (r *redirect) Status() int {
	return r.code
}

func Redirect(req *Request, url string, code int) Response {
	return &redirect{
		request: req.Request,
		url:     url,
		code:    code,
	}
}

func RedirectStage(url string, code int) Stage {
	return func(req *Request) (Response, error) {
		return Redirect(req, url, code), nil
	}
}

type createdResponse struct {
	location    string
	contentType string
	body        io.Reader
}

func (r *createdResponse) Render(w http.ResponseWriter) error {
	header := w.Header()
	if r.location != "" {
		header.Set("Location", r.location)
	}
	if r.contentType != "" {
		header.Set("Content-Type", r.contentType)
	}
	w.WriteHeader(http.StatusCreated)
	if r.body != nil {
		_, err := io.Copy(w, r.body)
		return err
	}
	return nil
}

func (r *createdResponse) Status() int {
	return http.StatusCreated
}

func Created(location string, contentType string, body io.ReadCloser) Response {
	return &createdResponse{location, contentType, body}
}

type notFoundResponse struct {
	body        io.Reader
	contentType string
}
