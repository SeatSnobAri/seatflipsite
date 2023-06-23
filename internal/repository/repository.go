package repository

import "github.com/SeatSnobAri/seatflipsite/internal/models"

type DatabaseRepo interface {
	// users and authentication
	GetUserById(id string) (models.User, error)
	InsertUser(u models.User) (int, error)
	Authenticate(email, testPassword string) (int, string, error)
	AllUsers() ([]*models.User, error)
	InsertRememberMeToken(id int, token string) error
	CheckForToken(id int, token string) bool
	AddUser(u models.GoogleUserResult) error

	// cart info
	InsertCart(payload models.UflipPayload, user models.UserPayload) error
	UpdateCart(buy bool, id string) error
	GetCartUser(id string) string
}
