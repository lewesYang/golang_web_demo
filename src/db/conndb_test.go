package db

import (
	"testing"
)

type User struct {
	Name  string
	Phone string
}

func Test_CreateSession(t *testing.T) {
	sess := CreateSession("127.0.0.1:27017")
	defer sess.Close()
	sess.DB("BLOG").C("USER").Insert(&User{Name: "lewesyang", Phone: "110"})
}
