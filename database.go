package w3sql

import (
	"database/sql"
	"errors"
	"reflect"
)

var structFieldsByNum = make(map[string]map[string]int)

// ScanToStruct scan row to s
func ScanToStruct(s interface{}, rows *sql.Rows) error {
	if s == nil {
		return errors.New("dest is <nil>")
	}

	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return errors.New("dest is not pointer")
	}

	el := v.Elem()
	t := el.Type()
	fields, ok := structFieldsByNum[t.Name()]
	if !ok {
		fields = prepareStructTags(t)
	}

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	var dest = make([]interface{}, len(cols))
	for i, col := range cols {
		fieldIdx, ok := fields[col]
		if ok {
			dest[i] = el.Field(fieldIdx).Addr().Interface()
			continue
		}

		var temp interface{}
		dest[i] = &temp
	}

	return rows.Scan(dest...)
}

func prepareStructTags(t reflect.Type) map[string]int {
	name := t.Name()
	fields := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		fields[t.Field(i).Tag.Get("db")] = i
	}

	structFieldsByNum[name] = fields
	return fields
}
