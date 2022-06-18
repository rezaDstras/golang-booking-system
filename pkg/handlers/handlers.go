package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rezaDastrs/pkg/config"
	"github.com/rezaDastrs/pkg/models"
	"github.com/rezaDastrs/pkg/render"
)

//the repository used by handlers
var Repo *Repository

//repository type
type Repository struct {
	App *config.AppConfig
}

//create new repository
func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

//set the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {

	//get Ip address
	remoteIp := r.RemoteAddr
	//put in session
	m.App.Session.Put(r.Context(), "remote_ip", remoteIp)

	//preform some logic

	stringMap := make(map[string]string)
	stringMap["test"] = "Hello World!"
	//send data to the template

	render.RenderTempl(w, "home.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//get ip address from session
	remoteIp := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap := make(map[string]string)
	stringMap["remote_ip"] = remoteIp

	render.RenderTempl(w, "About.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func Sum(x, y int) int {
	return x + y
}

func Divide(w http.ResponseWriter, r *http.Request) {
	var x float32
	var y float32
	x = 100.0
	y = 10.0

	f, err := divdeFunction(x, y)

	if err != nil {
		fmt.Fprintf(w, "can not divide by ziro")
		return
	}
	fmt.Fprintf(w, "%f divide by %f is %f", x, y, f)
}

func divdeFunction(x, y float32) (float32, error) {
	if y <= 0 {
		err := errors.New("can not divide by ziro")
		return 0, err
	}

	return x / y, nil

}
