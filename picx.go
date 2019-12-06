package main

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type page struct {
	//Id    bson.ObjectId `json:"id" bson:"_id,omitempty"`
	HtmlId        string
	Url           string `json:"group,omitempty" bson:",omitempty"`
	Title         string
	handler       func(http.ResponseWriter, *http.Request)
	Display       bool
	CustomContent string
}
type content struct {
	Pages []page
}

type message struct {
	Success bool
	Message string
	Path    string `json:",omitempty" bson:",omitempty"`
}

type user struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Username string
	Password string
}

//var pageContent = &content{}

var funcs = template.FuncMap{"isset": isset, "noescape": noescape}
var t, err = template.New("Picx").Funcs(funcs).ParseFiles("templates/index.html")

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	loggedIn := cookie != nil
	pages := getContent(loggedIn)
	fmt.Println("index handler")
	fmt.Println(loggedIn)
	fmt.Println(cookie)

	t.ExecuteTemplate(w, "index.html", pages)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("users")
	err = db.Insert(bson.M{
		"username": r.PostFormValue("username"),
		"password": r.PostFormValue("password"),
	})
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			response, _ := json.Marshal(message{Success: false, Message: "Name bereits vergeben"})
			w.Write(response)
		}
		return
	}
	response, _ := json.Marshal(message{Success: true, Message: "Registrierung erfolgreich"})
	w.Write(response)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("users")
	var user user
	err = db.Find(bson.M{
		"username": r.PostFormValue("username"),
	}).One(&user)
	w.Header().Set("Content-Type", "application/json")
	if err != nil { //len(users) == 0 {
		response, _ := json.Marshal(message{Success: false, Message: "User nicht gefunden"})
		w.Write(response)
		return
	}
	if user.Password == r.PostFormValue("password") {
		response, _ := json.Marshal(message{Success: true, Message: "Willkommen an Bord " + user.Username, Path: "login"})
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   user.Id.Hex(),
			Expires: time.Now().Add(500 * time.Second),
		})
		w.Write(response)
		return
	}
	response, _ := json.Marshal(message{Success: false, Message: "falsches Passwort"})
	w.Write(response)
}

func getContent(loggedIn bool) content {
	register := page{
		HtmlId:  "register-link",
		Title:   "Registrieren",
		Display: !loggedIn,
	}
	login := page{
		HtmlId:  "login-link",
		Title:   "Login",
		Display: !loggedIn,
	}
	logout := page{
		HtmlId:  "logout-link",
		Title:   "Logout",
		Display: loggedIn,
	}
	baseImages := page{
		HtmlId:  "base-images-link",
		Title:   "Basismotive",
		Display: loggedIn,
	}
	pools := page{
		HtmlId:  "pools-link",
		Title:   "Kachelpools",
		Display: loggedIn,
	}
	mosaics := page{
		HtmlId:  "mosaics-link",
		Title:   "Mosaike",
		Display: loggedIn,
	}
	deleteAccount := page{
		HtmlId:  "delete-account-link",
		Title:   "Account löschen",
		Display: loggedIn,
	}

	pageContent := content{
		[]page{
			register,
			login,
			logout,
			baseImages,
			pools,
			mosaics,
			deleteAccount,
		},
	}
	return pageContent
}

func isset(name string, data interface{}) bool {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}
	return v.FieldByName(name).IsValid()
}
func noescape(str string) template.HTML {
	return template.HTML(str)
}

func main() {
	//getContent()
	//pages := &pageContent.Pages
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)

	//for _, page := range pageContent.Pages {
	//	if isset("Url", page) && page.Url != "" && isset("handler", page) && page.handler != nil {
	//		http.HandleFunc("/"+page.Url, page.handler)
	//	}
	//}
	http.ListenAndServe(":4242", nil)
}
