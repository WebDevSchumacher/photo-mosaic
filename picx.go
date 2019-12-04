package main

import (
	"html/template"
	"log"
	"net/http"
	"reflect"
)

type page struct {
	//Id    bson.ObjectId `json:"id" bson:"_id,omitempty"`
	HtmlId        string
	Url           string
	Title         string
	handler       func(http.ResponseWriter, *http.Request)
	Display       bool
	CustomContent string
}
type content struct {
	Pages []page
}

var funcs = template.FuncMap{"isset": isset, "noescape": noescape}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, error := template.New("Picx").Funcs(funcs).ParseFiles("templates/index.html")
	if error != nil {
		log.Fatal(error)
	}
	t.ExecuteTemplate(w, "index.html", getContent())
}

func registerHandler(w http.ResponseWriter, r *http.Request) {

}

func loginHandler(w http.ResponseWriter, r *http.Request) {

}

func someHandler(w http.ResponseWriter, r *http.Request) {

}

func getContent() content {
	register := page{
		HtmlId:  "register-link",
		Title:   "Registrieren",
		handler: registerHandler,
		Display: true,
		CustomContent: `<div id="register-modal" class="modal">
  <div class="modal-content">
    <span class="close">&times;</span>
    <p>Some text in the Modal..</p>
  </div>
</div>`,
	}
	login := page{
		HtmlId:  "login-link",
		Url:     "login",
		Title:   "Login",
		Display: true,
	}
	somePage := page{
		Url:     "somepage",
		Title:   "some Page",
		handler: someHandler,
		Display: false,
	}
	content := content{
		[]page{
			register,
			login,
			somePage,
		}}
	return content
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
	pages := getContent().Pages
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.HandleFunc("/", indexHandler)

	for _, page := range pages {
		if isset("Url", page) && page.Url != "" && isset("handler", page) && page.handler != nil {
			http.HandleFunc("/"+page.Url, page.handler)
		}
	}

	http.ListenAndServe(":4242", nil)
}
