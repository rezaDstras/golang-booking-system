package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rezaDastrs/internal/config"
	"github.com/rezaDastrs/internal/driver"
	"github.com/rezaDastrs/internal/forms"
	"github.com/rezaDastrs/internal/helpers"
	"github.com/rezaDastrs/internal/models"
	"github.com/rezaDastrs/internal/render"
	"github.com/rezaDastrs/internal/repository"
	"github.com/rezaDastrs/internal/repository/dbrepo"
)

//the repository used by handlers
var Repo *Repository

//repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

//create new repository
func NewRepo(a *config.AppConfig, db *driver.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewPostgresRepo(&db.SQL, a),
	}
}

//set the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {

	render.Templ(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (m *Repository) About(w http.ResponseWriter, r *http.Request) {

	//get ip address from session
	remoteIp := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap := make(map[string]string)
	stringMap["remote_ip"] = remoteIp

	render.Templ(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	//get reservation key in session
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can not get reservation from session"))
		return
	}

	log.Println(res.RoomID)

	//convert dates recieved from session
	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	//create map to pass addetional value to tmpl
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	//get information from rom with room id which is passd in (reservation) session
	room, err := m.DB.GetRommByID(res.RoomID)
	if err != nil {
		helpers.ServerError(w, err)
	}

	//Room is property in reservation model (append room name to res)
	res.Room.RoomName = room.RoomName

	//update reservation key session with room name
	m.App.Session.Put(r.Context(), "reservation", res)

	data := make(map[string]interface{})
	data["reservation"] = res
	render.Templ(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      forms.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {

	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("can not get from session"))
		return
	}

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	//already in reservation session we have start date , end date , room id and room name
	reservation.FirstName = r.Form.Get("first_name")
	reservation.LastName = r.Form.Get("last_name")
	reservation.Email = r.Form.Get("email")
	reservation.Phone = r.Form.Get("phone")

	form := forms.New(r.PostForm)

	//form.Has("first_name", r)
	form.Required("first_name", "last_name", "email")
	form.MinLenght("first_name", 2, r)
	form.MinLenght("last_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation

		render.Templ(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	//insert data to db
	newReservationId, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//already in reservation session we have start date , end date , room id and room name
	restriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationId,
		RestictionID:  1,
	}

	err = m.DB.InsertRoomrestriction(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//put results in session which is defined in first line of main.go
	m.App.Session.Put(r.Context(), "resevation", reservation)

	//Send Email Notification

	htmlMsg := fmt.Sprintf(`
	<strong>Reservation Confirmation</strong><br>
	Dear : %s <br>
	This is a Confirmation for your reservation from %s to %s.
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))
	msg := models.MailData{
		To:      reservation.Email,
		From:    "you@here.com",
		Subject: "Reservation Confirmation",
		Content: htmlMsg,
	}

	// pass msg to config
	m.App.MailChan <- msg

	//redirect after submit form
	http.Redirect(w, r, "/reservation-summery", http.StatusSeeOther)

}

func (m *Repository) Availibility(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (m *Repository) PostAvailibility(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, end)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := m.DB.SearchAvailabilityForAllRooms(startDate, endDate)

	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// for _, i := range rooms {
	// 	m.App.InfoLog.Println("Rooms : ", i.Id, i.RoomName)
	// }

	if len(rooms) == 0 {
		//if no rooms available => return back with erro message
		m.App.Session.Put(r.Context(), "error", "No rooms available")
		http.Redirect(w, r, "/search-availability", http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	//save reservation dates available in session for using them in other pages
	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	m.App.Session.Put(r.Context(), "reservation", res)

	render.Templ(w, r, "choose-rooms.page.tmpl", &models.TemplateData{
		Data: data,
	})

	// w.Write([]byte(fmt.Sprintf("start date is %s and end date is %s", start, end)))
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (m *Repository) AvailibilityJSON(w http.ResponseWriter, r *http.Request) {
	resp := jsonResponse{
		OK:      true,
		Message: "Available!!",
	}

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByDateByRoomID(startDate, endDate, roomID)

	resp = jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (m *Repository) ReservationSummery(w http.ResponseWriter, r *http.Request) {
	//grab data from session which is set in PostReservation in session
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		m.App.ErrorLog.Println("can not get item from session")
		m.App.Session.Put(r.Context(), "error", "can not get item from session")
		//redirect to home page
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	//remove reservation session
	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["resrervation"] = reservation

	//format date in session
	sd := reservation.StartDate.Format("2006-01-02")
	ed := reservation.EndDate.Format("2006-01-02")

	//create map for pass formated date to tmpl
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed
	render.Templ(w, r, "reservation-summery.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (m *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	//get room id from guery
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//get reservation key wihich is saved in seassion
	res, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, err)
		return
	}
	// add room id to created seestion which name is reservation which is created in PostAvailibility function
	res.RoomID = roomID
	//update reservation key in session
	m.App.Session.Put(r.Context(), "reservation", res)
	//now we hame start date , end date and room id in reservation key in session

	//redirect to reservation page
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

//BookRoom takes URL parameters , Build a sessional variable , and takes users to make reservation tmpl
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	//get data from url query (id , s , e)
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	room, err := m.DB.GetRommByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.Room.RoomName = room.RoomName
	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	//put data to reservation session
	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)

}

func (m *Repository) LoginPage(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

func (m *Repository) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	//remove token in session
	_ = m.App.Session.RenewToken(r.Context())

	//parse form
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}

	//create form validate
	form := forms.New(r.PostForm)
	form.Required("email")
	form.Required("password")
	form.IsEmail("email")
	if !form.Valid() {
		render.Templ(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	id, _, err := m.DB.Authenticate(email, password)

	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "inavild login crediontials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "flash", "logged in successfully")
	m.App.Session.Put(r.Context(), "user_id", id)
	http.Redirect(w, r, "/", http.StatusSeeOther)

}

func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.Destroy(r.Context())
	_ = m.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (m *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.Templ(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

//Admin Panle
func (m *Repository) AdminAllReservations(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.Templ(w, r, "admin-all-reservation.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminNewReservations(w http.ResponseWriter, r *http.Request) {

	reservations, err := m.DB.NewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.Templ(w, r, "admin-new-reservation.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

func (m *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	//exploded url
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
	}

	src := exploded[3]

	stringMap := make(map[string]string)
	stringMap["src"] = src

	//get reservation by id

	reservation, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
	}

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.Templ(w, r, "admin-show-reservation.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
		Form:      forms.New(nil),
	})
}

func (m *Repository) AdminPostShowReservation(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//exploded url
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	//get reservation by id
	res, err := m.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//fill reservation data
	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	//update reservation

	err = m.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	//flash
	m.App.Session.Put(r.Context(), "flash", "changed saved!")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)

}

func (m *Repository) AdminProcessedReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.UpdateProcessedForReservation(id, 1)

	m.App.Session.Put(r.Context(), "flash", "Reservation marked  as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")

	_ = m.DB.DeleteReservation(id)

	m.App.Session.Put(r.Context(), "flash", "Reservation marked  as processed")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
}

func (m *Repository) AdminCalendarReservation(w http.ResponseWriter, r *http.Request) {

	// current time
	now := time.Now()

	//if url has parameter about last or first year / month
	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		//overwrite current time
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}
	//NEXT MONTH add 1 month
	next := now.AddDate(0, 1, 0)
	//previous month mines 1 month
	last := now.AddDate(0, -1, 0)

	//format next month and year
	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")
	//format last month & year
	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear

	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	data := make(map[string]interface{})
	data["now"] = now

	//number of days in current month
	//without return day _
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)

	//for access to last day of monh =>add a month to curent monh and mine a day
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	//Get All Rooms
	rooms, err := m.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	//determine which days of month which room si block or empty for reservation
	for _, x := range rooms {
		//create empty reservation days of rooms
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		//initialize days of month with 0
		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		//get all restrictions for current room
		restrictions, err := m.DB.GetRestrictionsForRoomByDate(x.Id, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}
		//check if is reserved or not
		for _, y := range restrictions {
			if y.ReservationID > 0 {
				//it is a reservation
				for d := y.StartDate; d.After(y.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = y.ReservationID
				}
			} else {
				//it is a bklock
				blockMap[y.StartDate.Format("2006-01-2")] = y.Id
			}
		}
		data[fmt.Sprintf("reservation_map_%d", x.Id)] = reservationMap
		data[fmt.Sprintf("block_map_%d", x.Id)] = blockMap

		//save block map for current month in session
		m.App.Session.Put(r.Context(),fmt.Sprintf("block_map_%d",x.Id),blockMap)
	}
	render.Templ(w, r, "admin-calendar-reservation.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}
