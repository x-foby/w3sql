package source

import "testing"

func TestDBName(t *testing.T) {
	s := &Source{
		Cols: NewCols(
			NewCol(TypeObject, "colA", "col_a", true).WithChildren(NewCols(
				NewCol(TypeString, "b", "b", false),
			)),
			NewCol(TypeString, "b", "b", true),
		),
	}

	if name, ok := s.Cols.DBName("colA.b"); !ok {
		t.Errorf("expected ok: %v, got: %v", true, false)
		t.Fail()
	} else if name != "col_a,b" {
		t.Errorf("expected name: %v, got: %v", "col_a,b", name)
		t.Fail()
	}
}
