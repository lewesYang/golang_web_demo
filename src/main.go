package main

import (
	"./session"
	"db"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io"
	"log"
	"model"
	"net/http"
	"net/url"
	"os/user"
	"strings"
	"time"
)

var mux map[string]func(http.ResponseWriter, *http.Request)
var homeTempl = template.Must(template.ParseFiles("view/chat.html"))

type Person struct {
	Name string
}

func main() {

	fmt.Println("listen 7777")
	u, _ := user.Current()
	fmt.Print(u.HomeDir)
	mux = make(map[string]func(http.ResponseWriter, *http.Request))
	mux["/"] = index
	mux["/out"] = signOut
	mux["/login"] = toIn
	mux["/in"] = signIn
	mux["/chat"] = serveChat
	mux["/ws"] = serveWs

	err := http.ListenAndServe(":7777", &myHandler{})
	if err != nil {
		log.Fatal(err)
	}
}

type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("go_session_id")
	if err != nil || cookie == nil {
		fmt.Println("no cookie or error")
		if r.URL.String() == "/in" {
			if h, ok := mux[r.URL.String()]; ok {
				h(w, r)
				return
			}
		}
		toIn(w, r)
		return
	}
	sessionMannger := session.SM
	ssid, _ := url.QueryUnescape(cookie.Value)
	fmt.Println("client_cookie:", ssid)
	se := sessionMannger.Get(ssid)
	fmt.Println("server_cookie_value:", se)
	if se == nil && r.URL.String() != "/in" {
		toIn(w, r)
	} else if h, ok := mux[r.URL.String()]; ok {
		h(w, r)
	} else if strings.HasPrefix(r.URL.String(), "/static/") {
		had := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
		had.ServeHTTP(w, r)
	} else {
		http.Error(w, "404 not found", 404)
	}
}
func toIn(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("view/login.html")
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

func signOut(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Bye bye, this is version 3.")
}

func signIn(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	fmt.Println("form_name:", name)
	db_session := db.CreateSession("127.0.0.1:27017")
	defer db_session.Close()
	result := model.User{}
	conn := db_session.DB("BLOG").C("USER")
	conn.Find(bson.M{"name": name}).One(&result)
	fmt.Println("userName:", result.Name)
	if result.Name == "" {
		toIn(w, r)
	} else {
		sessionMannger := session.SM
		ssid := sessionMannger.SessionStart(w, r)
		fmt.Println("set_ssid:", ssid)
		session := sessionMannger.NewSession(ssid, result) //创建session
		sessionMannger.SL.PushBack(session)                //存入session
		//io.WriteString(w, "wecome:"+name)
		index(w, r)
	}

}
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("view/index.html")
	if err != nil {
		fmt.Print(err)
		time.Sleep(5 * time.Second)
		log.Fatal(err)
	}
	t.Execute(w, nil)
}
