package models

import (
	"time"
)

type Reservation struct {
	Id        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
	Processed int
}

type User struct {
	Id          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Room struct {
	Id        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Restriction struct {
	Id              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type RoomRestriction struct {
	Id            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	RestictionID  int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReservationID int
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

type MailData struct {
	To string
	From string
	Subject string
	Content string
}
