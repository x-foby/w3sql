package w3sql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Server is a sources list
type Server struct {
	sources map[string]*Source
}

// NewServer return new Server
func NewServer() *Server {
	return &Server{
		sources: make(map[string]*Source),
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

// ServeHTTP is a default w3sql-handler
func (w3 Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	src, err := url.QueryUnescape(strings.Replace(r.URL.RequestURI(), "+", "$add$", -1))
	if err != nil {
		badRequest(w, err)
		return
	}
	src = strings.Replace(src, "$add$", "+", -1)

	var p Parser
	q, err := p.Parse(&w3, src)
	if err != nil {
		badRequest(w, err)
		return
	}

	path, method := q.Path(), r.Method

	s, ok := w3.sources[path]
	if !ok {
		notFound(w, method+" "+path)
		return
	}

	h, ok := s.Handlers[method]
	if !ok {
		notFound(w, method+" "+path)
		return
	}

	status, data, err := h(Context{&w3, q, w, r})
	if err != nil {
		printErr(w, status, err)
		return
	}

	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		printErr(w, status, err)
		return
	}

	w.WriteHeader(status)
	w.Write(buf)
}

func printErr(w http.ResponseWriter, code int, err error) {
	w.WriteHeader(code)
	w.Write([]byte(err.Error()))
}

func notFound(w http.ResponseWriter, address string) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(http.StatusText(http.StatusNotFound) + ": " + address))
}

func internalServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError) + ": " + err.Error()))
}

func badRequest(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(http.StatusText(http.StatusBadRequest) + ": " + err.Error()))
}

// func unescape(s string) string {

// }
