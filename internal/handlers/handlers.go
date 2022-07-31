package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/forms"
	"github.com/rezaDastrs/internal/models"
	"github.com/rezaDastrs/internal/render"
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

	render.RenderTempl(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//get ip address from session
	remoteIp := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap := make(map[string]string)
	stringMap["remote_ip"] = remoteIp

	render.RenderTempl(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTempl(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTempl(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation = models.Reservation{}
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation
	render.RenderTempl(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Panicln(err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	form := forms.New(r.PostForm)

	//form.Has("first_name", r)
	form.Required("first_name", "last_name", "enmail")
	form.MinLenght("first_name", 2, r)
	form.MinLenght("last_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.RenderTempl(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	//put results in session which is defined in first line of main.go
	m.App.Session.Put(r.Context(), "resevation", reservation)

	//redirect after submit form
	http.Redirect(w, r, "/reservation-summery", http.StatusOK)

}

func (m *Repository) Availibility(w http.ResponseWriter, r *http.Request) {
	render.RenderTempl(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailibility(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (m *Repository) AvailibilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available!!",
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTempl(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummery(w http.ResponseWriter, r *http.Request) {
	//grab data from session which is set in PostReservation in session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("can not get item from session")
		m.App.Session.Put(r.Context(), "error", "can not get item from session")
		//redirect to home page
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//remove session
	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["resrervation"] = reservation
	render.RenderTempl(w, r, "reservation-summery.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
