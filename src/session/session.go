package session

import (
	"container/list"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type SessionMannger struct {
	Lock       sync.Mutex //锁
	SL         *list.List //session 列表
	Expires    int        //有效期分钟
	cookieName string
}

type Session struct {
	Key     string
	Expires int
	Value   interface{}
}

var SM *SessionMannger

func (this *SessionMannger) NewSession(key string, value interface{}) *Session {
	return &Session{Key: key, Value: value, Expires: this.Expires}
}
func (this *SessionMannger) Listen() {
	fmt.Println(time.Now())
	time.AfterFunc(time.Second, func() { this.Listen() })
	this.Lock.Lock()
	defer this.Lock.Unlock()
	for ele := this.SL.Front(); ele != nil; ele = ele.Next() {
		fmt.Println("session_Expires:", ele.Value.(*Session).Expires)
		if ele.Value.(*Session).Expires == 0 {
			this.SL.Remove(ele)
			fmt.Println("session_time_out delete!")
		} else {
			ele.Value.(*Session).Expires--
		}
	}
}
func (this *SessionMannger) Get(key string) interface{} {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	for ele := this.SL.Front(); ele != nil; ele = ele.Next() {
		if key == ele.Value.(*Session).Key {
			ele.Value.(*Session).Expires = this.Expires
			this.SL.MoveToBack(ele)
			return ele.Value.(*Session).Value
		}
	}
	return nil
}

func (this *SessionMannger) Set(key string, value interface{}) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	for ele := this.SL.Front(); ele != nil; ele = ele.Next() {
		if key == ele.Value.(*Session).Key {
			this.SL.Remove(ele)
			break
		}
	}
	this.SL.PushBack(this.NewSession(key, value))
}

func (this *SessionMannger) Delete(key string) {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	for ele := this.SL.Front(); ele != nil; ele = ele.Next() {
		if key == ele.Value.(*Session).Key {
			this.SL.Remove(ele)
			return
		}
	}
}
func (this *SessionMannger) SessionStart(w http.ResponseWriter, r *http.Request) string {
	this.Lock.Lock()
	defer this.Lock.Unlock()
	ssid := this.SessionId()
	cookie := http.Cookie{Name: this.cookieName, Value: url.QueryEscape(ssid), Path: "/", HttpOnly: true, MaxAge: this.Expires * 1000}
	http.SetCookie(w, &cookie)
	return ssid
}
func (this *SessionMannger) SessionId() string {
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)

}
func init() {
	SM = &SessionMannger{SL: list.New(), Expires: 5000, cookieName: "go_session_id"}
	go SM.Listen()
}
