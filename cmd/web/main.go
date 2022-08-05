package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/driver"
	"github.com/rezaDastrs/internal/handlers"
	"github.com/rezaDastrs/internal/models"
	"github.com/rezaDastrs/internal/render"
)

const Port = ":8080"

var app config.AppConfig

var session *scs.SessionManager

func main() {
	db ,err := run()
	if err != nil {
		log.Fatal(err)
	}
	defer db.SQL.Close()


	fmt.Printf("starting application on prot %s", Port)
	// _ = http.ListenAndServe(Port, nil)

	srv := &http.Server{
		Addr:    Port,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() (*driver.DB , error){
	//define session 
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Reservation{})
	gob.Register(models.Restriction{})
	

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

	//conect to DB
	log.Println("connecting to DB")
	db ,err := driver.ConnectSQL("host=localhost port=5432 user=ehsan.d dbname=book password=")
	if err != nil {
		log.Fatal("connect to db error: ", err)
	}
	//template cache
	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("connot create template cache")
		return nil, err
	}

	//store the app
	app.TemplateCache = tc
	app.UseCache = false

	render.NewRenderer(&app)

	repo := handlers.NewRepo(&app ,db)
	handlers.NewHandlers(repo)
	
	return db , nil
}
