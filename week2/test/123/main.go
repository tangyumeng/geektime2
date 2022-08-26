package main

import (
	"fmt"
	"reflect"
)

func collectFieldNames(entity any, m *[]string, n *[]any) {

	if entity == nil {

	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()

	}
	num := typ.NumField()

	for i := 0; i < num; i++ {
		fd := typ.Field(i)
		fdVal := val.Field(i)
		if fdVal.Kind() == reflect.Struct {
			collectFieldNames(fdVal.Interface(), m, n)
		} else {
			fmt.Println("+++++", fd.Name)
			*m = append(*m, fd.Name)
			*n = append(*n, fdVal.Interface())
		}
		// if fdVal.Kind() == reflect.Struct {
		// 	collectFieldNames(fdVal, m, n)
		// } else {
		// 	*m = append(*m, fd.Name)
		// 	*n = append(*n, fdVal.Interface())
		// }
	}
}

type B struct {
	X string
	Y string
}

type D struct {
	B
	Z string
}

func main() {
	m := make([]string, 0)
	n := make([]any, 0)
	d := D{
		B: B{
			X: "XXXXX",
			Y: "YYYYYY",
		},
		Z: "ZZZZ",
	}
	collectFieldNames(d, &m, &n)
	// fmt.Println(m)
	for _, v := range m {
		fmt.Println(v)
	}

	fmt.Println("--------- n")

	for _, v := range n {
		fmt.Println(v)
	}
}
