//main func
//@author: baoqiang
//@time: 2021/10/26 19:56:34
package main

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xiaoaxe/7days-golang/axe-orm/axeorm"
	"github.com/xiaoaxe/7days-golang/axe-orm/axeorm/session"
)

func main() {
	e, err := axeorm.NewEngine("sqlite3", "my.db")
	if err != nil {
		panic(err)
	}
	defer e.Close()

	s := e.NewSession()
	s.Model(&Student{}).DropTable()

	_, err = e.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&Student{}).CreateTable()
		_, err = s.Insert(&Student{Id: 18, Name: "bbq"})
		return
	})

	var student Student
	err = s.Where("Id=?", 18).First(&student)
	fmt.Printf("Got Student: %v\n", student)
}

type Student struct {
	Id   int64 `axeorm:"PRIMARY KEY"`
	Name string
}
