package repository

import (
	"time"

	"github.com/rezaDastrs/internal/models"
)

type DatabaseRepo interface {
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomrestriction(r models.RoomRestriction) error
	SearchAvailabilityByDateByRoomID(start, end time.Time, roomId int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRommByID(id int) (models.Room, error)
	GetUserByID(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	//admin Panel
	AllReservations() ([]models.Reservation, error)
}
