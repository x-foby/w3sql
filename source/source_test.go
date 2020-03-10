package source

import "testing"

func TestJSONPath(t *testing.T) {
	s := &Source{
		Cols: NewCols(
			NewCol(TypeObject, "colA", "col_a", true).WithChildren(NewCols(
				NewCol(TypeObject, "b", "some_b", true).WithChildren(NewCols(
					NewCol(TypeString, "c", "c", false),
				)),
			)),
			NewCol(TypeString, "b", "b", true),
		),
	}

	if name, ok := s.Cols.JSONPath("colA.b.c"); !ok {
		t.Errorf("expected ok: %v, got: %v", true, false)
		t.Fail()
	} else if name != "some_b,c" {
		t.Errorf("expected name: %v, got: %v", "some_b,c", name)
		t.Fail()
	}
}
