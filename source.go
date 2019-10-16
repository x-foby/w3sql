package w3sql

// Datatype is a datatype
type Datatype int

// consts
const (
	Number Datatype = iota
	String
	Bool
	Time
	JSONArray
	JSONObject
)

// Common HTTP methods.
//
// Unless otherwise noted, these are defined in RFC 7231 section 4.3.
const (
	MethodGet     = "GET"
	MethodHead    = "HEAD"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodPatch   = "PATCH" // RFC 5789
	MethodDelete  = "DELETE"
	MethodConnect = "CONNECT"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
)

// Col is a column
type Col struct {
	Type     Datatype
	Name     string
	Required bool
}

// Cols is a columns map
type Cols map[string]Col

// A Source is a columns list
type Source struct {
	Cols     Cols
	Handlers map[string]Handler
	server   *Server
}

// NewSource return new Source
func NewSource(server *Server, cols Cols) *Source {
	return &Source{
		Cols:     cols,
		Handlers: make(map[string]Handler),
		server:   server,
	}
}

// Get is a handler for GET method
func (s *Source) Get(h Handler) *Source {
	return s.registerHandler(MethodGet, h)
}

// Head is a handler for HEAD method
func (s *Source) Head(h Handler) *Source {
	return s.registerHandler(MethodHead, h)
}

// Post is a handler for POST method
func (s *Source) Post(h Handler) *Source {
	return s.registerHandler(MethodPost, h)
}

// Put is a handler for PUT method
func (s *Source) Put(h Handler) *Source {
	return s.registerHandler(MethodPut, h)
}

// Patch is a handler for PATCH method
func (s *Source) Patch(h Handler) *Source {
	return s.registerHandler(MethodPatch, h)
}

// Delete is a handler for DELETE method
func (s *Source) Delete(h Handler) *Source {
	return s.registerHandler(MethodDelete, h)
}

// Connect is a handler for CONNECT method
func (s *Source) Connect(h Handler) *Source {
	return s.registerHandler(MethodConnect, h)
}

// Options is a handler for OPTIONS method
func (s *Source) Options(h Handler) *Source {
	return s.registerHandler(MethodOptions, h)
}

// Trace is a handler for TRACE method
func (s *Source) Trace(h Handler) *Source {
	return s.registerHandler(MethodTrace, h)
}

func (s *Source) registerHandler(m string, h Handler) *Source {
	s.Handlers[m] = h

	return s
}
