package binding

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

//
// SoftError A "soft" anticipated error.
type SoftError struct {
	Reason string
}

func (e *SoftError) Error() (s string) {
	return e.Reason
}

func (e *SoftError) Is(err error) (matched bool) {
	_, matched = err.(*SoftError)
	return
}

func (e *SoftError) Soft() *SoftError {
	return e
}

//
// RestError reports REST errors.
type RestError struct {
	SoftError
	Method string
	Path   string
	Status int
	Body   string
}

func (e *RestError) Is(err error) (matched bool) {
	_, matched = err.(*RestError)
	return
}

func (e *RestError) Error() (s string) {
	s = e.Reason
	return
}

func (e *RestError) With(r *http.Response) *RestError {
	e.Method = r.Request.Method
	e.Path = r.Request.URL.Path
	e.Status = r.StatusCode
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			e.Body = string(body)
		}
	}
	s := strings.ToUpper(e.Method)
	s += " "
	s += e.Path
	s += " failed: "
	s += strconv.Itoa(e.Status)
	s += "("
	s += http.StatusText(e.Status)
	s += ")"
	if e.Body != "" {
		s += " body: "
		s += e.Body
	}
	e.Reason = s
	return e
}

//
// Conflict reports 409 error.
type Conflict struct {
	RestError
}

func (e *Conflict) Is(err error) (matched bool) {
	_, matched = err.(*Conflict)
	return
}

//
// NotFound reports 404 error.
type NotFound struct {
	RestError
}

func (e *NotFound) Is(err error) (matched bool) {
	_, matched = err.(*NotFound)
	return
}
