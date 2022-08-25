package main

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type BaseEntity struct {
	CreateTime int64
	UpdateTime *int64
}

type User struct {
	BaseEntity
	Id       uint64
	NickName sql.NullString
	Age      *sql.NullInt32
}

var errInvalidEntity = errors.New("invalid entity")

func InsertStmt(entity interface{}) (string, []interface{}, error) {

	if entity == nil {
		return "", nil, errInvalidEntity
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	// 如果不是结构体，就返回 error
	if typ.Kind() != reflect.Struct {
		return "", nil, errInvalidEntity
	}

	num := typ.NumField()

	if num == 0 {
		return "", nil, errInvalidEntity
	}

	bd := strings.Builder{}
	bd.WriteString("INSERT INTO `")

	bd.WriteString(typ.Name())
	bd.WriteString("`(")

	res := []interface{}{}
	first := true
	for i := 0; i < num; i++ {
		fd := typ.Field(i)
		fdVal := val.Field(i)

		// if fd.Type.Kind() == reflect.String {
		// 	fmt.Println("Hello,world", fd.Type.Kind())
		// }

		localType := reflect.TypeOf(fd)

		if localType.Kind() == reflect.Struct {
			fmt.Println("检查出内嵌变量类型")
			localVal := reflect.ValueOf(fdVal)
			localFirst := true
			localNum := localType.NumField()

			for j := 0; j < localNum; j++ {
				localFd := localType.Field(j)
				localFdVal := localVal.Field(j)
				fmt.Println("0000")
				fmt.Println(localFdVal)
				res = append(res, localFdVal.Interface())
				fmt.Println("1111")
				if localFd.IsExported() {
					if localFirst {
						bd.WriteString("`")
						localFirst = false
					} else {
						bd.WriteString(",`")
					}
					bd.WriteString(localFd.Name)
					bd.WriteString("`")
				}
			}

		}

		fmt.Println("_________________++++++++++")
		fmt.Println(bd.String())
		fmt.Println("_________________++++++++++")

		res = append(res, fdVal.Interface())
		if fd.IsExported() {
			if first {
				bd.WriteString("`")
				first = false
			} else {
				bd.WriteString(",`")
			}
			bd.WriteString(fd.Name)
			bd.WriteString("`")
		}
	}

	bd.WriteString(") VALUES(")
	dotFirst := true
	for i := 0; i < num; i++ {
		if dotFirst {
			bd.WriteString("?")
			dotFirst = false
		} else {
			bd.WriteString(",?")
		}
	}

	bd.WriteString(");")
	fmt.Println("------------")
	fmt.Println(bd.String())
	fmt.Println(res...)
	fmt.Println("------------")

	// []interface{}{int64(0), (*int64)(nil)},

	return bd.String(), res, nil

	// val := reflect.ValueOf(entity)
	// typ := val.Type()
	// 检测 entity 是否符合我们的要求
	// 我们只支持有限的几种输入

	// 使用 strings.Builder 来拼接 字符串
	// bd := strings.Builder{}

	// 构造 INSERT INTO XXX，XXX 是你的表名，这里我们直接用结构体名字

	// 遍历所有的字段，构造出来的是 INSERT INTO XXX(col1, col2, col3)
	// 在这个遍历的过程中，你就可以把参数构造出来
	// 如果你打算支持组合，那么这里你要深入解析每一个组合的结构体
	// 并且层层深入进去

	// 拼接 VALUES，达成 INSERT INTO XXX(col1, col2, col3) VALUES

	// 再一次遍历所有的字段，要拼接成 INSERT INTO XXX(col1, col2, col3) VALUES(?,?,?)
	// 注意，在第一次遍历的时候我们就已经拿到了参数的值，所以这里就是简单拼接 ?,?,?

	// return bd.String(), args, nil
}

func main() {

	u := User{
		BaseEntity: BaseEntity{
			CreateTime: 123,
			UpdateTime: ptrInt64(456),
		},
		Id:       789,
		NickName: sql.NullString{String: "Tom", Valid: true},
	}

	i, j, k := InsertStmt(u)

	fmt.Println(i, j, k)
}

func ptrInt64(val int64) *int64 {
	return &val
}