package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/rezaDastrs/internal/models"
	"golang.org/x/crypto/bcrypt"
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

func (m *postgressDBRepo) GetRommByID(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room

	query := `
		select
			 id , room_name , created_at , updated_at
		from
			rooms
		where 
			id = $1	
	`

	res := m.DB.QueryRowContext(ctx, query, id)
	err := res.Scan(
		&room.Id,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *postgressDBRepo) GetUserByID(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			select 
				id , firest_name , last_name , email , password , access_level , created_at ,updated_at          
			from
				users
			where 
				id = $1
			`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.User

	err := row.Scan(
		&u.Id,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.AccessLevel,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *postgressDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			update
				 users
			set
				 firs_name = $1 , last_name = $2 , email = $3 , access_level = $4 , upadted_at = $5
	`

	_, err := m.DB.ExecContext(ctx, query,
		u.FirstName,
		u.LastName,
		u.Email,
		u.AccessLevel,
		time.Now(),
	)

	if err != nil {
		return nil
	}

	return nil
}

func (m *postgressDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
			select 
				id , password
			from
				users
			where
				email = $1			
	`

	var id int
	var hashPassword string

	row := m.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		//id
		&id,
		//password
		&hashPassword,
	)

	if err != nil {
		return id, "", err
	}

	//check pasword
	err = bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashPassword, nil

}

//Admin Panel

func (m *postgressDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
				select r.id, r.first_name, r.last_name, r.start_date, r.end_date, r.room_id, r.processed ,
				rb.id, rb.room_name 
				from reservations r
				left join rooms rb on (r.room_id = rb.id)		
				order by r.start_date asc								
			`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation

		err := rows.Scan(
			&i.Id,
			&i.FirstName,
			&i.LastName,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.Processed,
			&i.Room.Id,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)

	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

func (m *postgressDBRepo) NewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
				select r.id, r.first_name, r.last_name, r.start_date, r.end_date, r.room_id, r.processed ,
				rb.id, rb.room_name 
				from reservations r
				left join rooms rb on (r.room_id = rb.id)		
				where processed = 0 
				order by r.start_date asc								
			`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Reservation

		err := rows.Scan(
			&i.Id,
			&i.FirstName,
			&i.LastName,
			&i.StartDate,
			&i.EndDate,
			&i.RoomID,
			&i.Processed,
			&i.Room.Id,
			&i.Room.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		reservations = append(reservations, i)

	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

func (m *postgressDBRepo) GetReservationByID(id int) (models.Reservation, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `
			select r.id, r.first_name, r.last_name, r.phone, r.email, r.start_date, r.end_date, r.room_id, r.processed ,
				rb.id, rb.room_name 
			from reservations r
			left join rooms rb on (r.room_id = rb.id)	
			where r.id = $1		
	`

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&res.Id,
		&res.FirstName,
		&res.LastName,
		&res.Phone,
		&res.Email,
		&res.StartDate,
		&res.EndDate,
		&res.RoomID,
		&res.Processed,
		&res.Room.Id,
		&res.Room.RoomName,
	)

	if err != nil {
		return res, err
	}

	return res, nil
}

func (m *postgressDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set first_name = $1, last_name = $2, phone = $3, email = $4 , updated_at = $5 where id = $6`

	_, err := m.DB.ExecContext(ctx, query,
		r.FirstName,
		r.LastName,
		r.Phone,
		r.Email,
		time.Now(),
		r.Id,
	)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgressDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservations where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	return nil

}

func (m *postgressDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set processed = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil
}

func (m *postgressDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room

	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.Room

		err := rows.Scan(
			&i.Id,
			&i.RoomName,
			&i.CreatedAt,
			&i.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, i)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil

}

func (m *postgressDBRepo) GetRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var restrictions []models.RoomRestriction

	//if reservation id is null => give default 0 value

	query := `select id, coalesce (reservation_id, 0), restriction_id, room_id, start_date, end_date from room_restrictions where $1 < end_date and $2 > start_date and room_id = $3`

	rows, err := m.DB.QueryContext(ctx, query, end, start, roomID)
	if err != nil {
		return restrictions, err
	}

	defer rows.Close()

	for rows.Next() {
		var i models.RoomRestriction

		err := rows.Scan(
			&i.Id,
			&i.ReservationID,
			&i.RestictionID,
			&i.RoomID,
			&i.StartDate,
			&i.EndDate,
		)
		if err != nil {
			return restrictions, err
		}

		restrictions = append(restrictions, i)
	}
	if err = rows.Err(); err != nil {
		if err != nil {
			return restrictions, err
		}
	}
	return restrictions, nil
}
func (m *postgressDBRepo) InsertBlockForRoom(id int, start time.Time)  error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at ) values ($1, $2, $3, $4 , $5, $6)`

	_, err := m.DB.ExecContext(ctx, query, start, start.AddDate(0,0,1), id, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}
	return nil
}
func (m *postgressDBRepo) DeleteBlockForRoom(id int)  error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`


	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}