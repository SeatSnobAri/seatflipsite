package dbrepo

import (
	"context"
	"database/sql"
	"github.com/SeatSnobAri/seatflip2.0/internal/models"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

// AllUsers returns all users
func (repo *postgresDBRepo) AllUsers() ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT id, last_name, first_name, email, photo,verified,provider, created_at, updated_at FROM "users".users`

	rows, err := repo.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User

	for rows.Next() {
		s := &models.User{}
		err = rows.Scan(&s.ID, &s.LastName, &s.FirstName, &s.Email, &s.Photo, &s.Verified, &s.Provider, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Append it to the slice
		users = append(users, s)
	}

	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return users, nil
}

// GetUserById returns a user by id
func (repo *postgresDBRepo) GetUserById(id string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT id, last_name, first_name, email, photo, verified, provider, created_at, updated_at 
			FROM 
			    "users".users
		 	where 
		 	    id = $1`

	row := repo.DB.QueryRowContext(ctx, stmt, id)

	var u models.User

	err := row.Scan(
		&u.ID,
		&u.LastName,
		&u.FirstName,
		&u.Email,
		&u.Photo,
		&u.Verified,
		&u.Provider,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		log.Println(err)
		return u, err
	}

	return u, nil
}

// Authenticate authenticates
func (repo *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string
	var userActive int

	query := `
		select 
		    id, password, user_active 
		from 
			users 
		where 
			email = $1
			and deleted_at is null`

	row := repo.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword, &userActive)
	if err == sql.ErrNoRows {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		log.Println(err)
		return 0, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		log.Println(err)
		return 0, "", err
	}

	if userActive == 0 {
		return 0, "", models.ErrInactiveAccount
	}

	// Otherwise, the password is correct. Return the user ID and hashed password.
	return id, hashedPassword, nil
}

// InsertRememberMeToken inserts a remember me token into remember_tokens for a user
func (repo *postgresDBRepo) InsertRememberMeToken(id int, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "insert into remember_tokens (user_id, remember_token) values ($1, $2)"
	_, err := repo.DB.ExecContext(ctx, stmt, id, token)
	if err != nil {
		return err
	}
	return nil
}

// DeleteToken deletes a remember me token
func (repo *postgresDBRepo) DeleteToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "delete from remember_tokens where remember_token = $1"
	_, err := repo.DB.ExecContext(ctx, stmt, token)
	if err != nil {
		return err
	}

	return nil
}

// CheckForToken checks for a valid remember me token
func (repo *postgresDBRepo) CheckForToken(id int, token string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "SELECT id  FROM remember_tokens where user_id = $1 and remember_token = $2"
	row := repo.DB.QueryRowContext(ctx, stmt, id, token)
	err := row.Scan(&id)
	return err == nil
}

// Insert method to add a new record to the users table.
func (repo *postgresDBRepo) InsertUser(u models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	INSERT INTO "users".users 
	    (id,
		first_name, 
		last_name, 
		email, 
		photo, 
		verified,
		provider,
	    created_at,
	     updated_at
		)
    VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id `

	var newId int
	err := repo.DB.QueryRowContext(ctx, stmt,
		u.ID,
		u.FirstName,
		u.LastName,
		u.Email,
		u.Photo,
		u.Verified,
		u.Provider,
		time.Now(),
		time.Now()).Scan(&newId)
	if err != nil {
		return 0, err
	}

	return newId, err
}
func (repo *postgresDBRepo) AddUser(u models.GoogleUserResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	query := `insert into "users".users (id, first_name, last_name, email,photo,provider,verified, created_at, updated_at)
				values ($1,$2,$3,$4,$5,$6,$7,$8, $9)`
	log.Printf("%+v", u)
	_, err := repo.DB.ExecContext(ctx, query,
		&u.Id,
		&u.Given_name,
		&u.Family_name,
		&u.Email,
		&u.Picture,
		"google",
		&u.Verified_email,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return err
	}
	return nil
}
