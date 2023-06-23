package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SeatSnobAri/seatflipsite/internal/config"
	"github.com/SeatSnobAri/seatflipsite/internal/driver"
	"github.com/SeatSnobAri/seatflipsite/internal/forms"
	"github.com/SeatSnobAri/seatflipsite/internal/helpers"
	"github.com/SeatSnobAri/seatflipsite/internal/models"
	"github.com/SeatSnobAri/seatflipsite/internal/render"
	"github.com/SeatSnobAri/seatflipsite/internal/repository"
	"github.com/SeatSnobAri/seatflipsite/internal/repository/dbrepo"
	"github.com/SeatSnobAri/seatflipsite/internal/templates"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Repo is the repository
var Repo *DBRepo
var app *config.AppConfig

// DBRepo is the db repo
type DBRepo struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewHandlers creates the handlers
func NewHandlers(repo *DBRepo, a *config.AppConfig) {
	Repo = repo
	app = a
}

// NewPostgresqlHandlers creates db repo for postgres
func NewPostgresqlHandlers(db *driver.DB, a *config.AppConfig) *DBRepo {
	return &DBRepo{
		App: a,
		DB:  dbrepo.NewPostgresRepo(db.SQL, a),
	}
}

// Home is the home page handler
func (repo *DBRepo) Home(w http.ResponseWriter, r *http.Request) {
	render.Template(w, r, "home.page.gohtml", &templates.TemplateData{})
}

// Logout logs a user out
func (repo *DBRepo) Logout(w http.ResponseWriter, r *http.Request) {
	_ = repo.App.Session.Destroy(r.Context())
	_ = repo.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (repo *DBRepo) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	id := repo.App.Session.Get(r.Context(), "user_id")
	user, err := repo.DB.GetUserById(id.(string))
	if err != nil {
		log.Println(err)
		repo.App.Session.Put(r.Context(), "error", "can't get user from session")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// get all redis stuff
	rows, err := repo.sendCurrentRedis()
	if err != nil {
		log.Println(err)
		repo.App.Session.Put(r.Context(), "error", "can't get redis data")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := make(map[string]interface{})
	data["user"] = user
	data["rows"] = rows
	render.Template(w, r, "dashboard.page.gohtml", &templates.TemplateData{
		Data: data,
	})
}

func (repo *DBRepo) SignUp(w http.ResponseWriter, r *http.Request) {
	user := repo.App.Session.Get(r.Context(), "user")
	data := make(map[string]interface{})
	data["user"] = user

	render.Template(w, r, "sign-up.page.gohtml", &templates.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

func (repo *DBRepo) PostSignUp(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, r, err)
		return
	}
	verified, err := strconv.ParseBool(r.Form.Get("verified_email"))
	if err != nil {
		helpers.ServerError(w, r, err)
		return
	}

	user := models.GoogleUserResult{
		Given_name:     r.Form.Get("first_name"),
		Family_name:    r.Form.Get("last_name"),
		Email:          r.Form.Get("email"),
		Id:             r.Form.Get("id"),
		Picture:        r.Form.Get("picture"),
		Verified_email: verified,
	}

	form := forms.New(r.PostForm)

	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["user"] = user
		render.Template(w, r, "sign-up.page.gohtml", &templates.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}
	err = repo.DB.AddUser(user)
	if err != nil {
		log.Println(err)
		repo.App.Session.Put(r.Context(), "error", "problem adding user")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	repo.App.Session.Put(r.Context(), "user", user)
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (repo *DBRepo) Broker(w http.ResponseWriter, r *http.Request) {
	var jwt string
	user_id := app.Session.Get(r.Context(), "user_id")
	if user_id == "" {
		jwt = r.URL.Query().Get("token")
		v, _ := repo.validateGoogleJwt(jwt)
		if !v {
			helpers.ErrorJSON(w, errors.New("not valid google"))
			return
		}
	}

	var requestPayload models.RequestPayload

	err := helpers.ReadJSON(w, r, &requestPayload)
	if err != nil {
		helpers.ErrorJSON(w, err)
		return
	}

	switch requestPayload.Action {
	case "cart":
		repo.Produce(w, requestPayload.Produce, requestPayload.User)
	case "buy":
		repo.Consume(w, requestPayload.Buy)
	//case "va":
	//	repo.VABuy(w, requestPayload.VA)
	//case "delete":
	//	repo.Delete(w, requestPayload.Delete)
	default:
		helpers.ErrorJSON(w, errors.New("unknown action"))
	}
}
func (repo *DBRepo) Produce(w http.ResponseWriter, u models.UflipPayload, user models.UserPayload) {
	repo.storeInRedis(&u)

	err := repo.DB.InsertCart(u, user)
	if err != nil {
		helpers.ErrorJSON(w, err)
		return
	}
	data := make(map[string]string)

	data["buy"] = strconv.FormatBool(u.Buy)
	data["uuid"] = u.UUID
	data["event_date"] = u.EventDate
	data["event_name"] = u.EventName
	data["event_venue"] = u.EventVenue
	data["seat_info"] = u.SeatInfo
	data["ticket_info"] = u.TicketInfo
	data["ticket_price"] = u.TicketPrice
	data["ticket_total"] = u.TicketTotal
	data["stock_type"] = u.StockType

	repo.broadcastMessage("public-channel", "produce", data)

	var p models.JsonResponse
	p.Error = false
	p.Message = "we have now produced a cart item for your luxury if you would like to purchase its wonderful contents"

	helpers.WriteJSON(w, http.StatusAccepted, &p)
}

func (repo *DBRepo) Consume(w http.ResponseWriter, u models.BuyPayload) {
	var msg models.UflipPayload
	val, err := app.Redis.Get(context.Background(), u.UUID).Result()
	if err != nil {
		helpers.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	json.Unmarshal([]byte(val), &msg)
	msg.Buy = u.Buy
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	_ = app.Redis.Set(context.Background(), u.UUID, json, redis.KeepTTL)

	err = repo.DB.UpdateCart(true, u.UUID)
	if err != nil {
		helpers.ErrorJSON(w, err, http.StatusBadRequest)
		return
	}
	userId := repo.DB.GetCartUser(u.UUID)
	data := make(map[string]string)
	data["message"] = strconv.Itoa(msg.TabId)

	_ = repo.App.WsClient.Trigger("public-channel", fmt.Sprintf("%s", userId), data)

	var response models.JsonResponse
	response.Error = false
	response.Message = "updated Buy"

	helpers.WriteJSON(w, http.StatusAccepted, &response)
}

func (repo *DBRepo) storeInRedis(msg *models.UflipPayload) {
	json, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	if err = repo.App.Redis.Set(context.Background(), msg.UUID, json, 10*time.Minute).Err(); err != nil {
		panic(err)
	}
}
func (repo *DBRepo) sendCurrentRedis() ([]models.UflipPayload, error) {
	var cursor uint64
	var rows []models.UflipPayload
	iter := repo.App.Redis.Scan(context.Background(), cursor, "*", 10).Iterator()
	for iter.Next(context.Background()) {
		log.Println("keys", iter.Val())
		var msg models.UflipPayload
		m := repo.App.Redis.Get(context.Background(), iter.Val())
		json.Unmarshal([]byte(m.Val()), &msg)
		rows = append(rows, msg)

	}
	if err := iter.Err(); err != nil {
		panic(err)
	}
	return rows, nil
}
