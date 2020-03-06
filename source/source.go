package source

import "strings"

// Datatype is a datatype
type Datatype int

// consts
const (
	TypeNumber Datatype = iota
	TypeString
	TypeBool
	TypeTime
	TypeObject
)

// Col is a column
type Col struct {
	Type     Datatype
	IsArray  bool
	Children Cols
	Name     string
	Required bool
}

// NewCol returns new Col
func NewCol(datatype Datatype, name string, isArray bool) *Col {
	return &Col{
		Type:    datatype,
		Name:    name,
		IsArray: isArray,
	}
}

// WithChildren set a children cols for objects
func (c *Col) WithChildren(cols Cols) *Col {
	c.Children = cols
	return c
}

// Cols is a columns map
type Cols map[string]*Col

// NewCols returns new Col
func NewCols(cols ...*Col) Cols {
	return Cols{}.WithCols(cols...)
}

// NewCol add new col to cols
func (c Cols) NewCol(datatype Datatype, name string, isArray bool) Cols {
	c[name] = NewCol(datatype, name, isArray)
	return c
}

// WithCols add new col to cols
func (c Cols) WithCols(cols ...*Col) Cols {
	for _, col := range cols {
		if col != nil {
			c[col.Name] = col
		}
	}
	return c
}

// ByName returns column by name
func (c Cols) ByName(name string) *Col {
	return c.byName(name, c)
}

// Type returns columns datatype
func (c Cols) Type(name string) *Datatype {
	col := c.byName(name, c)
	if col == nil {
		return nil
	}
	return &col.Type
}

// ByName returns column by name
func (c Cols) byName(name string, cols Cols) *Col {
	parts := strings.Split(name, ".")
	key := name
	if len(parts) > 1 {
		key = parts[0]
	}
	col, ok := cols[key]
	if !ok {
		return nil
	}
	if len(parts) > 1 {
		n := strings.Join(parts[1:], ".")
		return c.byName(n, col.Children)
	}
	return col
}

// A Source is a columns list
type Source struct {
	Cols Cols
	// Handlers map[string]Handler
	// server   *Server
}
