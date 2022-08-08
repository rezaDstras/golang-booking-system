package render

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/models"
)

var functions = template.FuncMap{
	//for pass additional data like cuurent year or format of date or ...
	"humanDate": HumanDate,
	"formatDate": FormatDate,
	"iterate": Iterate,
	"add": Add,
}

var app *config.AppConfig

func HumanDate(t time.Time) string {
	return t.Format("02 jun 2006")
}

func Add(a ,b int) int {
	return a + b
}

//Iterate returns a slice of integers starting 1, going to count
func Iterate(count int) []int  {
	var i int
	var items []int

	for i = 1; i <= count; i++ {
		items = append(items, i)
	}

	return items
}


func FormatDate( t time.Time , f string) string{
	return t.Format(f)
}

//set the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	//show alert
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.CSRFToken = nosurf.Token(r)

	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}

	return td
}

func Templ(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	if app.UseCache {
		//get template cache from AppConfig
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	//get tmpl from map which has been defined in CreateTemplateCache
	t, ok := tc[tmpl]
	//check dosen't exist
	if !ok {
		//stop the application
		log.Fatal("could not get template from template cache")
	}

	//use without read from desk with buffer
	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)
	//hold information into bytes and read that from spesefic byte

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("error writing template to browser ", err)
	}
}

//create template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {

	//map for all template
	myChache := map[string]*template.Template{}

	//define location
	pages, err := filepath.Glob("./template/*.page.tmpl")

	if err != nil {
		return myChache, err
	}

	// foreach for all tmpl's

	for _, page := range pages {
		//get name of each page
		name := filepath.Base(page)

		// fmt.Println("page is currently :", page)

		// render tmpl file with using additional function in top

		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myChache, err
		}

		//get layout tmpl
		matches, err := filepath.Glob("./template/*.layout.tmpl")
		if err != nil {
			return myChache, err
		}

		//if find at lest ONE file it works, we check that with lenght
		if len(matches) > 0 {
			//combined page with layout
			ts, err = ts.ParseGlob("./template/*.layout.tmpl")
			if err != nil {
				return myChache, err
			}
		}

		//added to the myChache map
		myChache[name] = ts
	}

	return myChache, nil
}
