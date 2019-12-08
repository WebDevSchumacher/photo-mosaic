package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"html/template"
	"io"
	"log"
	"net/http"
	"reflect"
	"regexp"
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
	Pages              []page
	InnerContent       *template.Template
	InnerContentString string
}

type message struct {
	Success bool
	Message string
	Path    string `json:",omitempty" bson:",omitempty"`
}

type user struct {
	Id       bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Username string        `json:"username" bson:"username"`
	Password string        `json:"password" bson:"password"`
}

type baseBrowser struct {
}

type baseImageSet struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	Name   string        `json:"name" bson:"name"`
	Images []baseImage   `json:",omitempty" bson:",omitempty"`
}

type baseImageSets struct {
	Sets []baseImageSet
}

type baseImage struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	File   bson.ObjectId `json:"file" bson:"file"`
	Set    bson.ObjectId `json:"set" bson:"set"`
	Name   string        `json:"name" bson:"name"`
	Width  int           `json:",omitempty" bson:",omitempty"`
	Height int           `json:",omitempty" bson:",omitempty"`
}

var innerPageContent = &template.Template{}
var t *template.Template
var pageContent content
var funcs = template.FuncMap{"isset": isset, "noescape": noescape}

//var t, err = template.New("Picx").Funcs(funcs).ParseFiles("templates/index.html")
var routeMatch *regexp.Regexp

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	loggedIn := cookie != nil
	getContent(loggedIn)

	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "index_inner.html", nil)
	pageContent.InnerContentString = tpl.String()
	if loggedIn {
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   cookie.Value,
			Expires: time.Now().Add(500 * time.Second),
		})
	}

	w.WriteHeader(200)
	t.ExecuteTemplate(w, "index.html", pageContent)
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
	if err != nil {
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
func logoutHandler(w http.ResponseWriter, r *http.Request) {

}
func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {

}
func baseImagesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("baseImageSets")
	var sets []baseImageSet
	err = db.Find(bson.M{
		"user": bson.ObjectIdHex(cookie.Value),
	}).All(&sets)
	var setsCollection baseImageSets
	setsCollection.Sets = sets
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "base.html", setsCollection)
	response, _ := json.Marshal(message{Success: true, Message: tpl.String()})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(500 * time.Second),
	})
	w.Write(response)
}
func tilePoolsHandler(w http.ResponseWriter, r *http.Request) {

}
func mosaicsHandler(w http.ResponseWriter, r *http.Request) {

}
func newBaseImagesSetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("baseImageSets")
	err = db.Insert(bson.M{
		"name": r.PostFormValue("base-set-name"),
		"user": bson.ObjectIdHex(cookie.Value),
	})
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			response, _ := json.Marshal(message{Success: false, Message: "Name bereits vergeben"})
			w.Write(response)
		}
		return
	}
	response, _ := json.Marshal(message{Success: true, Message: "Set erfolgreich angelegt"})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(500 * time.Second),
	})
	w.Write(response)
}
func getBaseImagesSetHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	setId := r.URL.Query().Get("setId")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249")
	setsDb := db.C("baseImageSets")
	imagesDb := db.C("baseImagesMeta")
	var set baseImageSet
	err = setsDb.FindId(bson.ObjectIdHex(setId)).One(&set)
	err = imagesDb.Find(bson.M{
		"set": bson.ObjectIdHex(setId),
	}).All(&set.Images)

	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "setbrowser.html", set)
	response, _ := json.Marshal(message{Success: true, Message: tpl.String()})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(500 * time.Second),
	})
	w.Write(response)
}
func uploadBaseImagesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	if r.Method == "POST" && cookie != nil {
		reader, _ := r.MultipartReader()
		//r.ParseMultipartForm(32 << 20)
		//fmt.Println(r.MultipartForm)
		//fmt.Println(reader)
		session, err := mgo.Dial("localhost")
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()
		db := session.DB("Picx_ASchumacher_630249")
		baseImagesMeta := db.C("baseImagesMeta")
		baseImagesGridFs := db.GridFS("baseImages")
		fileName := ""
		fileExtension := ""
		setId := r.FormValue("set-id")
		fmt.Println(setId)
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if part.FormName() == "set-id" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(part)
				setId = buf.String()
				fmt.Println(setId)
			}
			if part.FileName() == "" {
				continue
			}
			fileName = part.FileName()
			if strings.Contains(fileName, ".") {
				split2 := strings.Split(fileName, ".")
				fileExtension = split2[len(split2)-1]
			}
			fileExtension = strings.ToLower(fileExtension)
			if fileExtension != "jpg" && fileExtension != "jpeg" && fileExtension != "png" {
				continue
			}
			gridFile, err := baseImagesGridFs.Create(fileName)
			_, err = io.Copy(gridFile, part)
			err = gridFile.Close()
			err = baseImagesMeta.Insert(bson.M{
				"user": bson.ObjectIdHex(cookie.Value),
				"file": gridFile.Id(),
				"set":  bson.ObjectIdHex(setId),
				"name": fileName,
			})
		}
		//type baseImage struct {
		//	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
		//	User   bson.ObjectId `json:"user" bson:"user"`
		//	File   bson.ObjectId `json:"file" bson:"file"`
		//	Set    bson.ObjectId `json:"set" bson:"set"`
		//	Name   string        `json:"name" bson:"name"`
		//	Width  int           `json:",omitempty" bson:",omitempty"`
		//	Height int           `json:",omitempty" bson:",omitempty"`
		//}
		response, _ := json.Marshal(message{Success: true, Message: "erfolgreich hochgeladen"})
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   cookie.Value,
			Expires: time.Now().Add(500 * time.Second),
		})
		w.Write(response)
	} else {
		indexHandler(w, r)
	}
}

func getContent(loggedIn bool) { //content {
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
	//innerPageContentLocal := template.New("empty")
	pageContent = content{
		[]page{
			register,
			login,
			logout,
			baseImages,
			pools,
			mosaics,
			deleteAccount,
		},
		innerPageContent,
		"",
	}
	//return pageContent
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
	var err error
	t, err = template.New("").Funcs(funcs).ParseGlob("./templates/*")
	//t, err = template.ParseGlob("./templates/*")
	//t.Funcs(funcs)
	if err != nil {
		log.Fatal(err)
	}
	routeMatch, _ = regexp.Compile(`^\/(\w+)`)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/delete-account", deleteAccountHandler)
	http.HandleFunc("/base-images", baseImagesHandler)
	http.HandleFunc("/base-images/new-set", newBaseImagesSetHandler)
	http.HandleFunc("/base-images/get-set", getBaseImagesSetHandler)
	http.HandleFunc("/base-images/upload", uploadBaseImagesHandler)
	http.HandleFunc("/tile-pools", tilePoolsHandler)
	http.HandleFunc("/mosaics", mosaicsHandler)
	http.ListenAndServe(":4242", nil)
}
