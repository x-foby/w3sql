package webserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/x-foby/w3sql/ast"
	"github.com/x-foby/w3sql/parser"
	"github.com/x-foby/w3sql/query"
	"github.com/x-foby/w3sql/source"
)

// Context contains Query and Request per every http-request
type Context struct {
	server *Server
	Query  *query.Query
	W      http.ResponseWriter
	R      *http.Request
}

// Handler is a callback that will be call per every http-request
type Handler func(ctx Context) (int, interface{}, error)

// Option is a servers option
type Option int

// options
const (
	OptJSONResult = iota + 1
	OptPrettyJSON
)

// SourceHandlers contains source and handlers
type SourceHandlers struct {
	Source   *source.Source
	Handlers map[string]Handler
}

// NewSourceHandlers return new NewSourceHandlers
func NewSourceHandlers(s *source.Source) *SourceHandlers {
	return &SourceHandlers{
		Source:   s,
		Handlers: make(map[string]Handler),
	}
}

// Get is a handler for GET method
func (s *SourceHandlers) Get(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodGet, h)
}

// Head is a handler for HEAD method
func (s *SourceHandlers) Head(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodHead, h)
}

// Post is a handler for POST method
func (s *SourceHandlers) Post(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodPost, h)
}

// Put is a handler for PUT method
func (s *SourceHandlers) Put(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodPut, h)
}

// Patch is a handler for PATCH method
func (s *SourceHandlers) Patch(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodPatch, h)
}

// Delete is a handler for DELETE method
func (s *SourceHandlers) Delete(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodDelete, h)
}

// Connect is a handler for CONNECT method
func (s *SourceHandlers) Connect(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodConnect, h)
}

// Options is a handler for OPTIONS method
func (s *SourceHandlers) Options(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodOptions, h)
}

// Trace is a handler for TRACE method
func (s *SourceHandlers) Trace(h Handler) *SourceHandlers {
	return s.registerHandler(http.MethodTrace, h)
}

func (s *SourceHandlers) registerHandler(m string, h Handler) *SourceHandlers {
	s.Handlers[m] = h
	return s
}

// Server is a sources list
type Server struct {
	resultAsJSON bool
	prettyJSON   bool
	errorHandler func(status int, err error) []byte
	sources      map[string]*SourceHandlers
}

// NewServer return new Server
func NewServer(options ...Option) *Server {
	return &Server{
		resultAsJSON: contains(options, OptJSONResult),
		sources:      make(map[string]*SourceHandlers),
	}
}

func unableRegister(msg string) error {
	return errors.New("unable to register source: " + msg)
}

// Route add Source into sources list
func (w3 *Server) Route(path string, s *SourceHandlers) error {
	if s == nil {
		return unableRegister("source handlers is <nil>")
	}
	if s.Source == nil {
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
	w3.serveHTTP(w, r, nil)
}

// ServeHTTPWithGlobals is a default w3sql-handler with global idents
func (w3 *Server) ServeHTTPWithGlobals(w http.ResponseWriter, r *http.Request, globals map[string]ast.Expr) {
	w3.serveHTTP(w, r, globals)
}

func (w3 *Server) serveHTTP(w http.ResponseWriter, r *http.Request, globals map[string]ast.Expr) {
	src, err := url.QueryUnescape(strings.Replace(r.URL.RequestURI(), "+", "$add$", -1))
	if err != nil {
		w3.error(w, http.StatusBadRequest, err)
		return
	}
	src = strings.Replace(src, "$add$", "+", -1)

	p := parser.New()
	q, err := p.Parse(src)
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

	status, data, err := h(Context{w3, q.WithSource(s.Source), w, r})
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
