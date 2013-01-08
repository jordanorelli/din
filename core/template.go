package din

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jordanorelli/din/dinutil"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var TemplateRoot string

var templateCache map[string]cachedTemplate

type cachedTemplate struct {
	*template.Template
	os.FileInfo
}

var ProjectRoot = ""

// exposes a shell command to the template layer.  Useful for debugging, but
// please don't rely on this stuff in production.
func shExpose(s string) func() (string, error) {
	return func() (string, error) {
		cmd := exec.Command(s)
		cmd.Dir = ProjectRoot
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

// gets the short hash of the current git project.  This is kinda volatile,
// since it's really dependent on the linked binary being in sync with the code
// on file.  I'm not a big fan of this strategy; the value should be computed
// once at compile time and included in the binary, but I don't know how to do
// that.
var gitShortHash = func() func() (string, error) {
	var hash string
	return func() (string, error) {
		if ProjectRoot == "" {
			panic("project root isn't set")
		}
		if hash == "" {
			cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
			cmd.Dir = ProjectRoot
			b, err := cmd.CombinedOutput()
			if err != nil {
				return "", err
			}
			hash = strings.Trim(string(b), "\n\r\t ")
		}
		return hash, nil
	}
}()

var templateFuncs = template.FuncMap{
	"jsonblob": func(v interface{}) (template.JS, error) {
		b, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return template.JS(b), nil
	},
	"csrf_token": func() (template.JS, error) {
		return template.JS(dinutil.RandomString(40)), nil
	},
	"csrf_key": func() (template.JS, error) {
		return template.JS(CSRFKey), nil
	},
	"git_shorthash": func() (string, error) {
		return gitShortHash()
	},
	"env":  shExpose("printenv"),
	"id":   shExpose("id"),
	"pwd":  shExpose("pwd"),
	"date": shExpose("date"),
}

type TemplateResponse struct {
	*template.Template
	Context    interface{}
	StatusCode int
}

func NewTemplateResponse(relpath string, context interface{}, code int) (*TemplateResponse, error) {
	t, err := Template(relpath)
	if err != nil {
		return nil, err
	}
	return &TemplateResponse{t, context, code}, nil
}

func (t *TemplateResponse) Render(w http.ResponseWriter) error {
	var buf bytes.Buffer
	if err := t.Template.Execute(&buf, t.Context); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func (t *TemplateResponse) Status() int {
	return t.StatusCode
}

func RegisterTemplateFn(key string, fn interface{}) {
	templateFuncs[key] = fn
}

func readTemplateFile(abspath string) (*template.Template, error) {
	relpath, err := filepath.Rel(TemplateRoot, abspath)
	if err != nil {
		return nil, fmt.Errorf(`din: unable to resolve template file path %v`, abspath)
	}

	t := template.New(relpath).Funcs(templateFuncs)

	b, err := ioutil.ReadFile(abspath)
	if err != nil {
		return nil, err
	}
	s := string(b)

	t, err = t.Parse(s)
	if err != nil {
		return nil, fmt.Errorf(`din: unable to read template file at path %v: %v`, abspath, err)
	}

	fi, err := os.Stat(abspath)
	if err != nil {
		return nil, fmt.Errorf(`din: unable to stat (2) template at path %v: %v`, abspath, err)
	}

	templateCache[relpath] = cachedTemplate{t, fi}
	return t, nil
}

func Template(relpath string) (*template.Template, error) {
	abspath := filepath.Join(TemplateRoot, relpath)
	if cached, ok := templateCache[relpath]; ok {
		fi, err := os.Stat(abspath)
		if err != nil {
			return nil, fmt.Errorf(`din: unable to stat template at path %v: %v`, abspath, err)
		}
		if fi.ModTime().After(cached.ModTime()) {
			return readTemplateFile(abspath)
		}
		return cached.Template, nil
	}
	return readTemplateFile(abspath)
}

func (r *Router) Template(pattern string, relpath string, code int) {
	var lastRetry time.Time
	t, err := Template(relpath)
	name := fmt.Sprintf(`template("%v")`, relpath)
	r.AddRoute(pattern, name, func(req *Request) (Response, error) {
		retry := func() {
			if time.Since(lastRetry) < time.Second {
				return
			}
			t, err = Template(relpath)
			lastRetry = time.Now()
		}
		retry()
		if err != nil {
			req.LogError(err)
			return nil, InternalServerError("unable to read template")
		}
		return &TemplateResponse{t, nil, code}, nil
	})
}

func init() {
	templateCache = make(map[string]cachedTemplate, 5)
}
