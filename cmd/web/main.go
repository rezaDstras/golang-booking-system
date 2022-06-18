package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/rezaDastrs/pkg/config"
	"github.com/rezaDastrs/pkg/handlers"
	"github.com/rezaDastrs/pkg/render"
)

const Port = ":8080"

var app config.AppConfig

var session *scs.SessionManager

func main() {
	//change this to true in production
	app.InProduction = false

	//session
	session = scs.New()
	session.Lifetime = 24 * time.Hour // 24 hours
	//store session in cookie
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	//development mood
	session.Cookie.Secure = app.InProduction

	app.Session = session

	//template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("connot create template cache")
	}

	//store the app
	app.TemplateCache = tc
	app.UseCache = false

	render.NewTemplates(&app)

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	// http.HandleFunc("/", handlers.Repo.Home)
	// http.HandleFunc("/about", handlers.About)
	// http.HandleFunc("/divide", handlers.Divide)

	fmt.Printf("starting application on prot %s", Port)
	// _ = http.ListenAndServe(Port, nil)

	srv := &http.Server{
		Addr:    Port,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}
