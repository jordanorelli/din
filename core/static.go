package din

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var StaticRoot = ""

func (r *Router) WhiteList(relpath string) {
	if r.staticWhitelist == nil {
		r.staticWhitelist = []string{relpath}
	} else {
		r.staticWhitelist = append(r.staticWhitelist, relpath)
	}
}

func (r *Router) TryStatic(w http.ResponseWriter, req *Request) error {
	if StaticRoot == "" {
        return InternalServerError("static root isn't set")
	}
	for _, p := range r.staticPaths {
		if p.MatchString(req.Request.URL.Path) {
			return ServeFile(w, req.Request, filepath.Join(StaticRoot, p.relpath))
		}
	}
	return ServeFile(w, req.Request, StaticRoot+req.Request.URL.Path)
}

type staticPath struct {
	regexp.Regexp
	relpath string
}

func (r *Router) Static(pattern string, relpath string) {
	p := regexp.MustCompile(pattern)
	if r.staticPaths == nil {
		r.staticPaths = make([]staticPath, 0, 10)
	}
	r.staticPaths = append(r.staticPaths, staticPath{Regexp: *p, relpath: relpath})
}

/* -----------------------------------------------------------------------------
*
*  everything below here is forked from the standard library.  Changes needed:
*
*    - disable directory listing
*    - expose 404 errors on serving static files
*
----------------------------------------------------------------------------- */

// ServeFile replies to the request with the contents of the named file or directory.
func ServeFile(w http.ResponseWriter, r *http.Request, name string) error {
	dir, file := filepath.Split(name)
	return serveFile(w, r, http.Dir(dir), file, false)
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

// if name is empty, filename is unknown. (used for mime type, before sniffing)
// if modtime.IsZero(), modtime is unknown.
// content must be seeked to the beginning of the file.
func serveContent(w http.ResponseWriter, r *http.Request, name string, modtime time.Time, size int64, content io.ReadSeeker) {
	if checkLastModified(w, r, modtime) {
		return
	}

	code := http.StatusOK

	// If Content-Type isn't set, use the file's extension to find it.
	if w.Header().Get("Content-Type") == "" {
		ctype := mime.TypeByExtension(filepath.Ext(name))
		if ctype == "" {
			// read a chunk to decide between utf-8 text and binary
			var buf [1024]byte
			n, _ := io.ReadFull(content, buf[:])
			b := buf[:n]
			ctype = http.DetectContentType(b)
			_, err := content.Seek(0, os.SEEK_SET) // rewind to output whole file
			if err != nil {
				http.Error(w, "seeker can't seek", http.StatusInternalServerError)
				return
			}
		}
		w.Header().Set("Content-Type", ctype)
	}

	// handle Content-Range header.
	// TODO(adg): handle multiple ranges
	sendSize := size
	if size >= 0 {
		ranges, err := parseRange(r.Header.Get("Range"), size)
		if err == nil && len(ranges) > 1 {
			err = errors.New("multiple ranges not supported")
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
			return
		}
		if len(ranges) == 1 {
			ra := ranges[0]
			if _, err := content.Seek(ra.start, os.SEEK_SET); err != nil {
				http.Error(w, err.Error(), http.StatusRequestedRangeNotSatisfiable)
				return
			}
			sendSize = ra.length
			code = http.StatusPartialContent
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ra.start, ra.start+ra.length-1, size))
		}

		w.Header().Set("Accept-Ranges", "bytes")
		if w.Header().Get("Content-Encoding") == "" {
			w.Header().Set("Content-Length", strconv.FormatInt(sendSize, 10))
		}
	}

	w.WriteHeader(code)

	if r.Method != "HEAD" {
		if sendSize == -1 {
			io.Copy(w, content)
		} else {
			io.CopyN(w, content, sendSize)
		}
	}
}

// name is '/'-separated, not filepath.Separator.
func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, name string, redirect bool) error {
	const indexPage = "/index.html"

	// redirect .../index.html to .../
	// can't use Redirect() because that would make the path absolute,
	// which would be a problem running under StripPrefix
	if strings.HasSuffix(r.URL.Path, indexPage) {
		localRedirect(w, r, "./")
		return nil
	}

	f, err := fs.Open(name)
	if err != nil {
		return Error{
			StatusCode: http.StatusNotFound,
			Message:    "file not found",
		}
	}
	defer f.Close()

	d, err1 := f.Stat()
	if err1 != nil {
		return Error{
			StatusCode: http.StatusInternalServerError,
			Message:    "unable to stat file",
		}
	}

	if redirect {
		// redirect to canonical path: / at end of directory url
		// r.URL.Path always begins with /
		url := r.URL.Path
		if d.IsDir() {
			if url[len(url)-1] != '/' {
				localRedirect(w, r, path.Base(url)+"/")
				return nil
			}
		} else {
			if url[len(url)-1] == '/' {
				localRedirect(w, r, "../"+path.Base(url))
				return nil
			}
		}
	}

	// use contents of index.html for directory, if present
	if d.IsDir() {
		if checkLastModified(w, r, d.ModTime()) {
			return nil
		}
		index := name + indexPage
		ff, err := fs.Open(index)
		if err == nil {
			defer ff.Close()
			dd, err := ff.Stat()
			if err == nil {
				name = index
				d = dd
				f = ff
			}
		}
	}

	if d.IsDir() {
		// dirList(w, f)
		return Error{
			StatusCode: http.StatusNotFound,
			Message:    "file not found",
		}
	}

	serveContent(w, r, d.Name(), d.ModTime(), d.Size(), f)
	return nil
}

// httpRange specifies the byte range to be sent to the client.
type httpRange struct {
	start, length int64
}

// parseRange parses a Range header string as per RFC 2616.
func parseRange(s string, size int64) ([]httpRange, error) {
	if s == "" {
		return nil, nil // header not present
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, errors.New("invalid range")
	}
	var ranges []httpRange
	for _, ra := range strings.Split(s[len(b):], ",") {
		i := strings.Index(ra, "-")
		if i < 0 {
			return nil, errors.New("invalid range")
		}
		start, end := ra[:i], ra[i+1:]
		var r httpRange
		if start == "" {
			// If no start is specified, end specifies the
			// range start relative to the end of the file.
			i, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				return nil, errors.New("invalid range")
			}
			if i > size {
				i = size
			}
			r.start = size - i
			r.length = size - r.start
		} else {
			i, err := strconv.ParseInt(start, 10, 64)
			if err != nil || i > size || i < 0 {
				return nil, errors.New("invalid range")
			}
			r.start = i
			if end == "" {
				// If no end is specified, range extends to end of the file.
				r.length = size - r.start
			} else {
				i, err := strconv.ParseInt(end, 10, 64)
				if err != nil || r.start > i {
					return nil, errors.New("invalid range")
				}
				if i >= size {
					i = size - 1
				}
				r.length = i - r.start + 1
			}
		}
		ranges = append(ranges, r)
	}
	return ranges, nil
}
