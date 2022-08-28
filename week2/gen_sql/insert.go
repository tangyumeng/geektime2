package homework

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var errInvalidEntity = errors.New("invalid entity")

func InsertStmt(entity interface{}) (string, []interface{}, error) {

	g := NewGenSQL()
	m := make([]string, 0)
	n := make([]any, 0)
	g.CollectFieldNames(entity, &m, &n)

	if g.Err != nil {
		return "", nil, g.Err
	}
	bd := strings.Builder{}
	bd.WriteString("INSERT INTO `")

	bd.WriteString(g.EntityName)
	bd.WriteString("`(")
	isfirst := true
	for _, v := range m {
		if isfirst {
			isfirst = false
			bd.WriteString("`")
			bd.WriteString(v)
			bd.WriteString("`")
		} else {
			bd.WriteString(",")
			bd.WriteString("`")
			bd.WriteString(v)
			bd.WriteString("`")
		}
	}

	bd.WriteString(") VALUES(")

	isfirst = true
	for _ = range m {
		if isfirst {
			isfirst = false
			bd.WriteString("?")
		} else {
			bd.WriteString(",")
			bd.WriteString("?")
		}
	}
	bd.WriteString(");")

	fmt.Println("result is:", bd.String())

	return bd.String(), n, nil
}

type GenSQL struct {
	Err        error
	EntityName string
}

func NewGenSQL() *GenSQL {
	return &GenSQL{}
}

func (g *GenSQL) CollectFieldNames(entity any, m *[]string, n *[]any) {

	if entity == nil {
		g.Err = errInvalidEntity
		return
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() == reflect.Ptr {
		g.Err = errInvalidEntity
		return
	}

	if g.EntityName == "" {
		g.EntityName = val.Type().Name()
	}

	num := typ.NumField()

	if num == 0 {
		g.Err = errInvalidEntity
		return
	}

	for i := 0; i < num; i++ {
		fd := typ.Field(i)
		fdVal := val.Field(i)

		_ = fd

		// fdValInterface := fdVal.Interface()
		isSql := fdVal.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem())
		embeded := val.Type().Field(i).Anonymous
		isStruct := val.Field(i).Kind() == reflect.Struct

		if embeded && isStruct && !isSql {
			g.CollectFieldNames(fdVal.Interface(), m, n)
		} else {
			*m = append(*m, fd.Name)
			*n = append(*n, fdVal.Interface())
		}
	}
}
