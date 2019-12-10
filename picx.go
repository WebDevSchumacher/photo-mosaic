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
	"math"
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
type baseImageSet struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	Name   string        `json:"name" bson:"name"`
	Images []baseImage   `json:",omitempty" bson:",omitempty"`
}
type tilePool struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	Name   string        `json:"name" bson:"name"`
	Images []tile        `json:",omitempty" bson:",omitempty"`
}
type mosaicCollection struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	Name   string        `json:"name" bson:"name"`
	Images []mosaic      `json:",omitempty" bson:",omitempty"`
}
type baseImageSets struct {
	Sets []baseImageSet
}
type tilePools struct {
	Pools []tilePool
}
type mosaicCollections struct {
	Collections []mosaicCollection
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
type tile struct {
	Id     bson.ObjectId `json:"id" bson:"_id,omitempty"`
	User   bson.ObjectId `json:"user" bson:"user"`
	File   bson.ObjectId `json:"file" bson:"file"`
	Pool   bson.ObjectId `json:"pool" bson:"pool"`
	Name   string        `json:"name" bson:"name"`
	Width  int           `json:",omitempty" bson:",omitempty"`
	Height int           `json:",omitempty" bson:",omitempty"`
	AvgR   uint8         `json:"avgr" bson:"avgr"`
	AvgG   uint8         `json:"avgg" bson:"avgg"`
	AvgB   uint8         `json:"avgb" bson:"avgb"`
	AvgA   uint8         `json:"avga" bson:"avga"`
}
type mosaic struct {
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
func tilesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("tilePools")
	var pools []tilePool
	err = db.Find(bson.M{
		"user": bson.ObjectIdHex(cookie.Value),
	}).All(&pools)
	var poolsCollection tilePools
	poolsCollection.Pools = pools
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "tiles.html", poolsCollection)
	response, _ := json.Marshal(message{Success: true, Message: tpl.String()})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
	})
	w.Write(response)
}

func mosaicsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("mosaicCollections")
	var collections []mosaicCollection
	err = db.Find(bson.M{
		"user": bson.ObjectIdHex(cookie.Value),
	}).All(&collections)
	var mosaicCollections mosaicCollections
	mosaicCollections.Collections = collections
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "mosaics.html", mosaicCollections)
	response, _ := json.Marshal(message{Success: true, Message: tpl.String()})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
	})
	w.Write(response)
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
func newTilePoolHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("tilePools")
	err = db.Insert(bson.M{
		"name": r.PostFormValue("tile-pool-name"),
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
func newMosaicCollectionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").C("mosaicCollections")
	err = db.Insert(bson.M{
		"name": r.PostFormValue("mosaic-collection-name"),
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
func getTilePoolHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	poolId := r.URL.Query().Get("poolId")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249")
	poolsDb := db.C("tilePools")
	tilesDb := db.C("tilesMeta")
	var pool tilePool
	err = poolsDb.FindId(bson.ObjectIdHex(poolId)).One(&pool)
	err = tilesDb.Find(bson.M{
		"pool": bson.ObjectIdHex(poolId),
	}).All(&pool.Images)
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "poolbrowser.html", pool)
	response, _ := json.Marshal(message{Success: true, Message: tpl.String()})
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   cookie.Value,
		Expires: time.Now().Add(time.Second * 60 * 60 * 24),
		Path:    "/",
	})
	w.Write(response)
}
func getMosaicCollectionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	collectionId := r.URL.Query().Get("collectionId")
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249")
	collectionsDb := db.C("mosaicCollections")
	imagesDb := db.C("mosaicsMeta")
	var collection mosaicCollection
	err = collectionsDb.FindId(bson.ObjectIdHex(collectionId)).One(&collection)
	err = imagesDb.Find(bson.M{
		"collection": bson.ObjectIdHex(collectionId),
	}).All(&collection.Images)
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "collectionbrowser.html", collection)
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

func uploadTilesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	if r.Method == "POST" && cookie != nil {
		reader, _ := r.MultipartReader()
		session, err := mgo.Dial("localhost")
		if err != nil {
			log.Fatal(err)
		}
		defer session.Close()
		db := session.DB("Picx_ASchumacher_630249")
		tilesMeta := db.C("tilesMeta")
		tilesGridFs := db.GridFS("tiles")
		fileName := ""
		fileExtension := ""
		poolId := r.FormValue("pool-id")
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if part.FormName() == "pool-id" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(part)
				poolId = buf.String()
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
			var buf bytes.Buffer
			tee := io.TeeReader(part, &buf)
			img, _, err := image.Decode(tee)
			avgR, avgG, avgB, avgA := averageColor(img)
			gridFile, err := tilesGridFs.Create(fileName)
			_, err = io.Copy(gridFile, &buf)
			err = gridFile.Close()
			err = tilesMeta.Insert(bson.M{
				"user": bson.ObjectIdHex(cookie.Value),
				"file": gridFile.Id(),
				"pool": bson.ObjectIdHex(poolId),
				"name": fileName,
				"avgr": avgR,
				"avgg": avgG,
				"avgb": avgB,
				"avga": avgA,
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

func loadBaseImageHandler(w http.ResponseWriter, r *http.Request) {
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

func loadTileHandler(w http.ResponseWriter, r *http.Request) {
	imageId := bson.ObjectIdHex(r.URL.Query().Get("image"))
	session, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	db := session.DB("Picx_ASchumacher_630249").GridFS("tiles")
	image, err := db.OpenId(imageId)
	w.Header().Add("Content-Type", "image/png")
	_, err = io.Copy(w, image)
	err = image.Close()
}

func testImageHandler(w http.ResponseWriter, r *http.Request) {
	src, _ := imaging.Open("./tiles-rescale/rescale0.jpg")
	fmt.Println(averageColor(src))

	//fmt.Println(uint8(avgR), uint8(avgG), uint8(avgB), uint8(avgA))
	//fmt.Println(math.Round(avgR), math.Round(avgG), math.Round(avgB), math.Round(avgA))
	//fmt.Println(uint8(math.Round(avgR)), uint8(math.Round(avgG)), uint8(math.Round(avgB)), uint8(math.Round(avgA)))
	//
	//fmt.Println(avgColor)
	//avgColor.RGBA()

	return
	files, err := ioutil.ReadDir("/home/andre/Downloads/webprogPicts/")
	if err != nil {
		log.Fatal(err)
	}
	//var src image.Image

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
	http.HandleFunc("/picx/tile-pools", tilesHandler)
	http.HandleFunc("/picx/tile-pools/new-pool", newTilePoolHandler)
	http.HandleFunc("/picx/tile-pools/get-pool", getTilePoolHandler)
	http.HandleFunc("/picx/tile-pools/upload", uploadTilesHandler)
	http.HandleFunc("/picx/mosaic-collections", mosaicsHandler)
	http.HandleFunc("/picx/mosaic-collections/new-collection", newMosaicCollectionHandler)
	http.HandleFunc("/picx/mosaic-collections/get-collection", getMosaicCollectionHandler)
	http.HandleFunc("/picx/load-base-image", loadBaseImageHandler)
	http.HandleFunc("/picx/load-tile", loadTileHandler)
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
		HtmlId:  "tile-pools-link",
		Title:   "Kachelpools",
		Display: loggedIn,
	}
	mosaics := page{
		HtmlId:  "mosaic-collections-link",
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

func averageColor(src image.Image) (uint8, uint8, uint8, uint8) {
	bounds := src.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y
	pixels := width * height
	fmt.Println(pixels)
	var pxr uint32
	var pxg uint32
	var pxb uint32
	var pxa uint32
	var avgR float64
	var avgG float64
	var avgB float64
	var avgA float64
	var count float64
	count = 1
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			pxr, pxg, pxb, pxa = src.At(x, y).RGBA()
			avgR += (float64(pxr>>8) - avgR) / count
			avgG += (float64(pxg>>8) - avgG) / count
			avgB += (float64(pxb>>8) - avgB) / count
			avgA += (float64(pxa>>8) - avgA) / count
			count++
		}
	}
	return uint8(math.Round(avgR)), uint8(math.Round(avgG)), uint8(math.Round(avgB)), uint8(math.Round(avgA))
}
