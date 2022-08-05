package dbrepo

import (
	"context"
	"time"

	"github.com/rezaDastrs/internal/models"
)

func (m *postgressDBRepo) InsertReservation(res models.Reservation) (int, error) {
	//use transaction  with 3s timeout => die after 3s
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var NewID int

	//returning id => pass id from creating item (like last id in maria db)
	stmt := `insert into reservations (first_name , last_name, email , phone ,start_date,end_date,room_id, created_at , updated_at)
			values($1 ,$2,$3,$4,$5,$6,$7,$8,$9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&NewID) //pass created id by QueryRowContext and scan it

	if err != nil {
		return 0, err
	}

	return NewID, nil
}

func (m *postgressDBRepo) InsertRoomrestriction(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `insert into room_restrictions (start_date,end_date,room_id,reservation_id,
		created_at,updated_at,restriction_id) values ($1,$2,$3,$4,$5,$6,$7)`

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestictionID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgressDBRepo) SearchAvailabilityByDateByRoomID(start, end time.Time, roomID int) (bool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	query := `
			select
				count(id)
			from
				room_restrictions	
			where
				room_id = $1
			and
				$2 < end_date and $3 > start_date;		
			`

	//QueryRowContext return row count
	row := m.DB.QueryRowContext(ctx, query,
		roomID,
		end,
		start,
	)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	}

	return false, nil
}

//return slice of rooms
func (m *postgressDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `
			select
				r.id , r.rome_name
			from	
				rooms r
			where
				r.id not in
			(
			select
				 room_id 
			from 
				 room_restrictions rr 
			where 
				$1 < rr.end_date and $2 > rr.start_date
			)			
	`

	rows, err := m.DB.QueryContext(ctx, query, end, start)

	if err != nil {
		return rooms, nil
	}

	for rows.Next() {
		var room models.Room

		err := rows.Scan(
			&room.Id,
			&room.RoomName,
		)
		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgressDBRepo) GetRommByID(id int) (models.Room , error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var  room models.Room

	query := `
		select
			 id , room_name , created_at , updated_at
		from
			rooms
		where 
			id = $1	
	`

	res := m.DB.QueryRowContext(ctx,query , id)
	err := res.Scan(
		&room.Id,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room , err
	}

	return room , nil
}
