package main

import (
	"fmt"
	"reflect"
	"strings"
)

type Users struct {
	Id     int
	Name   string
	Age    int
	Market map[int]string
	Source *Sfrom
	Ext    Info
}
type Info struct {
	Detail string
}
type Sfrom struct {
	Area string
}

func (u Users) Login() {
	fmt.Println("login")
}

func main() {
	m := map[int]string{1: "abc"}
	s := &Sfrom{Area: "beijing"}
	i := Info{Detail: "detail"}
	u := &Users{Id: 12, Market: m, Ext: i, Source: s}
	v := reflect.ValueOf(u)
	Explicit(v, 0)
}

func Explicit(v reflect.Value, depth int) {
	if v.CanInterface() {
		t := v.Type()
		switch v.Kind() {
		case reflect.Ptr:
			Explicit(v.Elem(), depth)
		case reflect.Struct:
			fmt.Printf(strings.Repeat("\t", depth)+"%v %v {\n", t.Name(), t.Kind())
			for i := 0; i < v.NumField(); i++ {
				f := v.Field(i)
				if f.Kind() == reflect.Struct || f.Kind() == reflect.Ptr {
					fmt.Printf(strings.Repeat("\t", depth+1)+"%s %s : \n", t.Field(i).Name, f.Type())
					Explicit(f, depth+2)
				} else {
					if f.CanInterface() {
						fmt.Printf(strings.Repeat("\t", depth+1)+"%s %s : %v \n", t.Field(i).Name, f.Type(), f.Interface())
					} else {

						fmt.Printf(strings.Repeat("\t", depth+1)+"%s %s : %v \n", t.Field(i).Name, f.Type(), f)
					}
				}
			}
			fmt.Println(strings.Repeat("\t", depth) + "}")
		}
	} else {
		fmt.Printf(strings.Repeat("\t", depth)+"%+v\n", v)
	}
}
