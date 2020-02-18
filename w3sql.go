package w3sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Option is a servers option
type Option int

// options
const (
	OptJSONResult = iota + 1
	OptPrettyJSON
)

// Server is a sources list
type Server struct {
	resultAsJSON bool
	prettyJSON   bool
	errorHandler func(status int, err error) []byte
	sources      map[string]*Source
}

// NewServer return new Server
func NewServer(options ...Option) *Server {
	return &Server{
		resultAsJSON: contains(options, OptJSONResult),
		sources:      make(map[string]*Source),
	}
}

func unableRegister(msg string) error {
	return errors.New("unable to register source: " + msg)
}

// Route add Source into sources list
func (w3 *Server) Route(path string, s *Source) error {
	if s == nil {
		return unableRegister("source is <nil>")
	}

	if _, ok := w3.sources[path]; ok {
		return unableRegister(fmt.Sprintf("path %s is already registered", path))
	}

	w3.sources[path] = s

	return nil
}

// SetErrorHandler set handler that returns []byte
func (w3 *Server) SetErrorHandler(errorHandler func(status int, err error) []byte) *Server {
	w3.errorHandler = errorHandler
	return w3
}

// ServeHTTP is a default w3sql-handler
func (w3 *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	src, err := url.QueryUnescape(strings.Replace(r.URL.RequestURI(), "+", "$add$", -1))
	if err != nil {
		w3.error(w, http.StatusBadRequest, err)
		return
	}
	src = strings.Replace(src, "$add$", "+", -1)

	var p Parser
	q, err := p.Parse(w3, src)
	if err != nil {
		w3.error(w, http.StatusBadRequest, err)
		return
	}

	path, method := q.Path(), r.Method

	s, ok := w3.sources[path]
	if !ok {
		w3.error(w, http.StatusNotFound, errors.New(method+" "+path))
		return
	}

	h, ok := s.Handlers[method]
	if !ok {
		w3.error(w, http.StatusNotFound, errors.New(method+" "+path))
		return
	}

	status, data, err := h(Context{w3, q, w, r})
	if err != nil {
		w3.error(w, status, err)
		return
	}

	buf, err := marshalJSON(data, w3.prettyJSON)
	if err != nil {
		w3.error(w, status, err)
		return
	}

	w.WriteHeader(status)
	w.Write(buf)
}

func (w3 *Server) error(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	if !w3.resultAsJSON {
		w.Write([]byte(err.Error()))
		return
	}

	if w3.errorHandler != nil {
		w.Write(w3.errorHandler(code, err))
		return
	}

	buf, err := marshalJSON(err, w3.prettyJSON)
	if err != nil {
		w.Write([]byte(http.StatusText(code)))
	}

	w.Write(buf)
}

func contains(options []Option, option Option) bool {
	for _, o := range options {
		if o == option {
			return true
		}
	}
	return false
}

func marshalJSON(data interface{}, pretty bool) ([]byte, error) {
	if pretty {
		return json.MarshalIndent(data, "", "  ")
	}
	return json.Marshal(data)
}
