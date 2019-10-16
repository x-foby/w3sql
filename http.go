package w3sql

import "net/http"

// Context contains Query and Request per every http-request
type Context struct {
	server *Server
	Query  *Query
	W      http.ResponseWriter
	R      *http.Request
}

// Handler is a callback that will be call per every http-request
type Handler func(ctx Context) (int, interface{}, error)
