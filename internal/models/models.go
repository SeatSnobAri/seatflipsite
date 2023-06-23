package models

import (
	"errors"
	"time"
)

var (
	// ErrNoRecord no record found in database error
	ErrNoRecord = errors.New("models: no matching record found")
	// ErrInvalidCredentials invalid username/password error
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	// ErrDuplicateEmail duplicate email error
	ErrDuplicateEmail = errors.New("models: duplicate email")
	// ErrInactiveAccount inactive account error
	ErrInactiveAccount = errors.New("models: Inactive Account")
)

// User model
type User struct {
	ID          string
	FirstName   string
	LastName    string
	AccessLevel int
	Photo       string
	Email       string
	Verified    bool
	Provider    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Preferences map[string]string
}
type GoogleUserResult struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Verified_email bool   `json:"verified_email"`
	Name           string `json:"name,omitempty"`
	Given_name     string `json:"give_name,omitempty"`
	Family_name    string `json:"family_name,omitempty"`
	Picture        string `json:"picture"`
	Locale         string `json:"locale,omitempty"`
}
type UflipPayload struct {
	TabId       int    `json:"tab_id"`
	StockType   string `json:"stock_type"`
	UUID        string `json:"uuid"`
	EventDate   string `json:"event_date"`
	EventName   string `json:"event_name"`
	EventVenue  string `json:"event_venue"`
	SeatInfo    string `json:"seat_info"`
	TicketInfo  string `json:"ticket_info"`
	TicketPrice string `json:"ticket_price"`
	TicketTotal string `json:"ticket_total"`
	Buy         bool   `json:"buy"`
}
type RequestPayload struct {
	Action  string        `json:"action"`
	Produce UflipPayload  `json:"cart,omitempty"`
	Buy     BuyPayload    `json:"buy,omitempty"`
	VA      VABuyPayload  `json:"va, omitempty"`
	Delete  DeletePayload `json:"delete,omitempty"`
	User    UserPayload   `json:"user"`
}
type UserPayload struct {
	Email string `json:"email"`
	Id    string `json:"id"`
}
type BuyPayload struct {
	Buy  bool   `json:"buy"`
	UUID string `json:"uuid"`
}
type VABuyPayload struct {
	RedisKey string `json:"key"`
}
type DeletePayload struct {
	RedisKey string `json:"key"`
}
type JsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}
