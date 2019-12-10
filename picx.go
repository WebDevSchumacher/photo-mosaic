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
	"log"
	"math"
	"net/http"
	"sort"
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
	Pages []page
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
	Id          bson.ObjectId      `json:"id" bson:"_id,omitempty"`
	User        bson.ObjectId      `json:"user" bson:"user"`
	Name        string             `json:"name" bson:"name"`
	Images      []baseImage        `json:",omitempty" bson:",omitempty"`
	Pools       []tilePool         `json:",omitempty" bson:",omitempty"`
	Collections []mosaicCollection `json:",omitempty" bson:",omitempty"`
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

var t *template.Template
var pageContent content
var session *mgo.Session
var err error

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, _ := r.Cookie("token")
	loggedIn := cookie != nil
	getContent(loggedIn)
	tpl := &bytes.Buffer{}
	t.ExecuteTemplate(tpl, "index_inner.html", nil)
	//pageContent.InnerContentString = tpl.String()
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
	var poolsCollection tilePools
	poolsCollection.Pools = getTilePoolsByUser(cookie.Value)
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
	var mosaicCollections mosaicCollections
	mosaicCollections.Collections = getMosaicCollections(cookie.Value)
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
	db := session.DB("Picx_ASchumacher_630249")
	setsDb := db.C("baseImageSets")
	imagesDb := db.C("baseImagesMeta")
	var set baseImageSet
	err = setsDb.FindId(bson.ObjectIdHex(setId)).One(&set)
	err = imagesDb.Find(bson.M{
		"set": bson.ObjectIdHex(setId),
	}).All(&set.Images)
	set.Pools = getTilePoolsByUser(cookie.Value)
	set.Collections = getMosaicCollections(cookie.Value)
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
	img := getBaseImage(r.URL.Query().Get("image"))
	w.Header().Add("Content-Type", "image/png")
	_, err := io.Copy(w, img)
	err = img.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func loadTileHandler(w http.ResponseWriter, r *http.Request) {
	imageId := bson.ObjectIdHex(r.URL.Query().Get("image"))
	db := session.DB("Picx_ASchumacher_630249").GridFS("tiles")
	img, _ := db.OpenId(imageId)
	w.Header().Add("Content-Type", "image/png")
	_, err = io.Copy(w, img)
	err = img.Close()
}

func createMosaicHandler(w http.ResponseWriter, r *http.Request) { //see upload handler
	cookie, _ := r.Cookie("token")
	if r.Method == "POST" && cookie != nil {
		fmt.Println("creating mosaic")

		tilePool := getTilePoolById(r.PostFormValue("tilePool")).Images

		file := getBaseImage(r.PostFormValue("baseId"))
		baseImg, _ := imaging.Decode(file)
		bounds := baseImg.Bounds()
		width := bounds.Max.X - bounds.Min.X
		height := bounds.Max.Y - bounds.Min.Y

		mosaic := imaging.New(width*20, height*20, color.RGBA{
			R: 0,
			G: 0,
			B: 0,
			A: 255,
		})
		type delta struct {
			Delta float64
			File  bson.ObjectId
		}
		var tiles []delta
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				pxR, pxG, pxB, _ := baseImg.At(x, y).RGBA()
				for i := 0; i < len(tilePool); i++ {
					tile := tilePool[0]
					dR := tile.AvgR - uint8(pxR>>8)
					dG := tile.AvgG - uint8(pxG>>8)
					dB := tile.AvgB - uint8(pxB>>8)
					d := math.Abs(math.Sqrt(float64(dR) + float64(dG) + float64(dB)))
					_ = append(tiles, delta{
						Delta: d,
						File:  tile.File,
					})
				}
				sort.SliceStable(tiles, func(i, j int) bool {
					return tiles[i].Delta < tiles[j].Delta
				})

			}
		}

		//avgR, avgG, avgB, avgA := averageColor(baseImg)

		fmt.Println(r.PostFormValue("baseId"))
		fmt.Println(r.PostFormValue("multiUseTile") != "")
		fmt.Println(r.PostFormValue("tilePool"))
		fmt.Println(r.PostFormValue("mosaicCollection"))
		fmt.Println(r.PostFormValue("nBest"))

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

func testImageHandler(w http.ResponseWriter, r *http.Request) {

	col := color.RGBA{
		R: 123,
		G: 234,
		B: 123,
		A: 255,
	}
	fmt.Println(col.R, col.G, col.B, col.A)

	//tilePool := getTilePoolById("5dee644a6b1553966096a797")
	//fmt.Println(tilePool)

	//for i := 0; i < 1; i++ {
	//	pict := imaging.New(20, 20, color.RGBA{
	//		R: 0,
	//		G: 0,
	//		B: 0,
	//		A: 255,
	//	})
	//	bounds := pict.Bounds()
	//	for x := bounds.Min.X; x < bounds.Max.X; x++ {
	//		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
	//			pict.Set(x, y, color.RGBA{
	//				R: uint8(rand.Uint32()),
	//				G: uint8(rand.Uint32()),
	//				B: uint8(rand.Uint32()),
	//				A: 255,
	//			} )
	//		}
	//	}
	//	imaging.Save(pict, "./tiles-rnd/rnd"+strconv.Itoa(i)+".png")
	//}

	//col := color.RGBA{
	//	R: 123,
	//	G: 123,
	//	B: 123,
	//	A: 255,
	//}

	//file := getBaseImage("5defa32703a5e93fdbee555b")
	//imgObj, err := imaging.Decode(file)
	//fmt.Println(imgObj.At(5,5))
	//file := getBaseImage("5defa16403a5e939438712a4")
	//imgObj, err := imaging.Decode(file)
	//fmt.Println(imgObj.At(5,5))
	//fmt.Printf("%T\n", imgObj)
	//return
	//buf := []byte{}
	//file.Read(buf)
	//fmt.Printf("%T\n", file)
	//imgObj, err := imaging.Decode(file)

	//for i := 0; i < 200; i++ {
	//
	//	fmt.Println(imgObj.At(200,i))
	//}
	//src, _ := imaging.Open("./tiles-rescale/rescale0.jpg")
	//fmt.Println(averageColor(src))

	//fmt.Println(uint8(avgR), uint8(avgG), uint8(avgB), uint8(avgA))
	//fmt.Println(math.Round(avgR), math.Round(avgG), math.Round(avgB), math.Round(avgA))
	//fmt.Println(uint8(math.Round(avgR)), uint8(math.Round(avgG)), uint8(math.Round(avgB)), uint8(math.Round(avgA)))
	//
	//fmt.Println(avgColor)
	//avgColor.RGBA()

	//files, err := ioutil.ReadDir("/home/andre/Downloads/webprogPicts/")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//var src image.Image

	//i := 0
	//expDir := "/home/andre/Studium/Sem5/Webprog/hausarbeit/tiles-rescale/"
	//for _, f := range files {
	//	name := f.Name()
	//	nameSplit := strings.Split(name, ".")
	//	if len(nameSplit) < 2 {
	//		continue
	//	}
	//	src, _ := imaging.Open("/home/andre/Downloads/webprogPicts/" + f.Name())
	//	resized := imaging.Resize(src, 20, 20, imaging.NearestNeighbor)
	//	imaging.Save(resized, expDir+"rescale"+strconv.Itoa(i)+"."+nameSplit[1])
	//	i++
	//}

	//fmt.Println(len(imaging.Histogram(src)))
	//src, _ := imaging.Open("blackheart.png")
	//fmt.Println(src.At(2, 2).RGBA())
	//
	//type SubImager interface {
	//	SubImage(r image.Rectangle) image.Image
	//}

	//var sub image.Image
	//for i := 0; i < 1540; i += 20 {
	//	for j := 0; j < 780; j += 20 {
	//
	//		sub = src.(SubImager).SubImage(image.Rect(i, j, i+20, j+20))
	//		imaging.Save(sub, "./pict/tile"+strconv.Itoa(i)+"-"+strconv.Itoa(j)+".png")
	//	}
	//}
	//
	//for i := 0; i < 1000; i++ {
	//	col := color.RGBA{
	//		R: uint8(rand.Uint32()),
	//		G: uint8(rand.Uint32()),
	//		B: uint8(rand.Uint32()),
	//		A: uint8(rand.Uint32()),
	//	}
	//	pict := imaging.New(20, 20, col)
	//	imaging.Save(pict, "./tiles/tile"+strconv.Itoa(i)+".png")
}

func main() {
	var err error
	session, err = mgo.Dial("localhost")
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()
	t, err = template.New("").ParseGlob("./templates/*")
	if err != nil {
		log.Fatal(err)
	}
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
	http.HandleFunc("/picx/create-mosaic", createMosaicHandler)
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
	}
}

func averageColor(src image.Image) (uint8, uint8, uint8, uint8) {
	bounds := src.Bounds()
	//width := bounds.Max.X - bounds.Min.X
	//height := bounds.Max.Y - bounds.Min.Y
	//pixels := width * height
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
func getTilePoolsByUser(userId string) []tilePool {
	db := session.DB("Picx_ASchumacher_630249").C("tilePools")
	var pools []tilePool
	err = db.Find(bson.M{
		"user": bson.ObjectIdHex(userId),
	}).All(&pools)
	return pools
}
func getMosaicCollections(userId string) []mosaicCollection {
	db := session.DB("Picx_ASchumacher_630249").C("mosaicCollections")
	var collections []mosaicCollection
	err = db.Find(bson.M{
		"user": bson.ObjectIdHex(userId),
	}).All(&collections)
	return collections
}
func getBaseImage(imageId string) *mgo.GridFile {
	db := session.DB("Picx_ASchumacher_630249").GridFS("baseImages")
	image, _ := db.OpenId(bson.ObjectIdHex(imageId))
	return image
}
func getTilePoolById(id string) tilePool {
	poolId := bson.ObjectIdHex(id)
	poolsDb := session.DB("Picx_ASchumacher_630249").C("tilePools")
	var pool tilePool
	err = poolsDb.FindId(poolId).One(&pool)
	tilesDb := session.DB("Picx_ASchumacher_630249").C("tilesMeta")
	err = poolsDb.FindId(poolId).One(&pool)
	err = tilesDb.Find(bson.M{
		"pool": poolId,
	}).All(&pool.Images)
	return pool
}
