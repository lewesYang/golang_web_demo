package session

import (
	"fmt"
	"testing"
)

func Test_SessionId(t *testing.T) {
	fmt.Println(SM.SessionId())
	t.Log(SM.SessionId())
}
