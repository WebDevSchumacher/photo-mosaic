package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"html/template"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type page struct {
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
			Expires: time.Now().Add(time.Second * 60 * 60 * 24),
			Path:    "/",
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
			Expires: time.Now().Add(time.Second * 60 * 60 * 24),
			Path:    "/",
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
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
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
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
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
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
	})
	w.Write(response)
}
func uploadBaseImagesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	if r.Method == "POST" && cookie != nil {
		reader, _ := r.MultipartReader()
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
		response, _ := json.Marshal(message{Success: true, Message: "erfolgreich hochgeladen"})
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   cookie.Value,
			Expires: time.Now().Add(time.Second * 60 * 60 * 24),
			Path:    "/",
		})
		w.Write(response)
	} else {
		indexHandler(w, r)
	}
}

func loadImageHandler(w http.ResponseWriter, r *http.Request) {
	imageId := bson.ObjectIdHex(r.URL.Query().Get("image"))
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").GridFS("baseImages")
	image, err := db.OpenId(imageId)
	w.Header().Add("Content-Type", "image/png")
	_, err = io.Copy(w, image)
	err = image.Close()
}

func testImageHandler(w http.ResponseWriter, r *http.Request) {

	files, err := ioutil.ReadDir("/home/andre/Downloads/webprogPicts/")
	if err != nil {
		log.Fatal(err)
	}
	var src image.Image
	i := 0
	expDir := "/home/andre/Studium/Sem5/Webprog/hausarbeit/tiles-rescale/"
	for _, f := range files {
		name := f.Name()
		nameSplit := strings.Split(name, ".")
		if len(nameSplit) < 2 {
			continue
		}
		src, _ := imaging.Open("/home/andre/Downloads/webprogPicts/" + f.Name())
		resized := imaging.Resize(src, 20, 20, imaging.NearestNeighbor)
		imaging.Save(resized, expDir+"rescale"+strconv.Itoa(i)+"."+nameSplit[1])
		i++
		//fmt.Println(i)
	}
	return

	//fmt.Println(len(imaging.Histogram(src)))
	//src, _ := imaging.Open("blackheart.png")
	fmt.Println(src.At(2, 2).RGBA())

	type SubImager interface {
		SubImage(r image.Rectangle) image.Image
	}

	var sub image.Image
	for i := 0; i < 1540; i += 20 {
		for j := 0; j < 780; j += 20 {

			sub = src.(SubImager).SubImage(image.Rect(i, j, i+20, j+20))
			imaging.Save(sub, "./pict/tile"+strconv.Itoa(i)+"-"+strconv.Itoa(j)+".png")
		}
	}

	for i := 0; i < 1000; i++ {
		col := color.RGBA{
			R: uint8(rand.Uint32()),
			G: uint8(rand.Uint32()),
			B: uint8(rand.Uint32()),
			A: uint8(rand.Uint32()),
		}
		pict := imaging.New(20, 20, col)
		imaging.Save(pict, "./tiles/tile"+strconv.Itoa(i)+".png")
	}
}

func main() {
	var err error
	t, err = template.New("").Funcs(funcs).ParseGlob("./templates/*")
	if err != nil {
		log.Fatal(err)
	}
	routeMatch, _ = regexp.Compile(`^\/(\w+)`)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/picx", indexHandler)
	http.HandleFunc("/picx/register", registerHandler)
	http.HandleFunc("/picx/login", loginHandler)
	http.HandleFunc("/picx/logout", logoutHandler)
	http.HandleFunc("/picx/delete-account", deleteAccountHandler)
	http.HandleFunc("/picx/base-images", baseImagesHandler)
	http.HandleFunc("/picx/base-images/new-set", newBaseImagesSetHandler)
	http.HandleFunc("/picx/base-images/get-set", getBaseImagesSetHandler)
	http.HandleFunc("/picx/base-images/upload", uploadBaseImagesHandler)
	http.HandleFunc("/picx/tile-pools", tilePoolsHandler)
	http.HandleFunc("/picx/mosaics", mosaicsHandler)
	http.HandleFunc("/picx/load-image", loadImageHandler)
	http.HandleFunc("/picx/test-image", testImageHandler)
	http.ListenAndServe(":4242", nil)
}
func getContent(loggedIn bool) {
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
