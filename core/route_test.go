package din

import "testing"

func TestNoMatch(t *testing.T) {
	r := NewRoute("^/foo$")
	if match := r.Match("/foo/bar"); match != nil {
		t.Errorf("found incorrect match")
	}
}

func TestExactMatch(t *testing.T) {
	r := NewRoute("^/foo$")
	match := r.Match("/foo")
	if match == nil {
		t.Errorf("match is %v, expected %v", match, nil)
	}
	if len(match.Args) > 0 {
		t.Errorf("expected no match args.  Found %v", match.Args)
	}
	if len(match.Kwargs) > 0 {
		t.Errorf("expected no match kwargs.  Found %v", match.Kwargs)
	}
}

// tests positional arguments in url path parsing
func TestArgs(t *testing.T) {
	r := NewRoute("^/foo/(.*)$")
	match := r.Match("/foo/bar")
	if match == nil {
		t.Errorf("expected match; found none.")
	}
	if match.Args[0] != "bar" {
		t.Errorf(`Found match args of %v, expected ["bar"]`, match.Args[0])
	}
	if len(match.Kwargs) != 0 {
		t.Errorf("Found some kwargs.  Shouldn't have.")
	}
}

// tests keyword arguments in url path parsing (i.e., named capture groups in
// the provided regex)
func TestKwargs(t *testing.T) {
	r := NewRoute("^/foo/(?P<leeroy>.*)$")
	match := r.Match("/foo/jenkins")
	if match == nil {
		t.Errorf("expected match; found none.")
	}
	if len(match.Args) > 0 {
		t.Errorf("Found some args.  Shouldn't have.")
	}
	if match.Kwargs == nil {
		t.Errorf("Got nil kwargs.  Shouldn't.")
	}
	if match.Kwargs["leeroy"] != "jenkins" {
		t.Errorf(`Found bad kwarg.  Expected "jenkins", found %v`, match.Kwargs["leeroy"])
	}
}

// tests a heterogenous mixture of positional and keyword parameters.
func TestHeterogenous(t *testing.T) {
	r := NewRoute("^/foo/(?P<leeroy>\\w+)/(\\d+)$")
	match := r.Match("/foo/jenkins/9000")
	if match == nil {
		t.Errorf("expected match; found none.")
	}
	if match.Args[0] != "9000" {
		t.Errorf(`expected arg of "9000", found %v`, match.Args[0])
	}
	if match.Kwargs["leeroy"] != "jenkins" {
		t.Errorf(`expected kwarg "leeroy" of value "jenkins".  Found %v`, match.Kwargs["leeroy"])
	}

	badmatch := r.Match("/foo/jenkins/")
	if badmatch != nil {
		t.Errorf("found bad match.")
	}
}
