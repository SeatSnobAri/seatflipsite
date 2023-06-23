package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/SeatSnobAri/seatflipsite/internal/helpers"
	"github.com/SeatSnobAri/seatflipsite/internal/models"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io"
	"log"
	"net/http"
	"time"
)

var googleOauthConfig = &oauth2.Config{
	//RedirectURL:  "https://seatflip.ninja/auth/google/callback",
	RedirectURL:  "http://localhost:4000/auth/google/callback",
	ClientID:     "46211357222-2tgfbaigpul4vn2hv1hp0v9v3n3sp8a4.apps.googleusercontent.com",
	ClientSecret: "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",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func (repo *DBRepo) OauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	err := repo.App.Session.RenewToken(r.Context())
	if err != nil {
		helpers.ServerError(w, r, err)
		return
	}

	// Create oauthState cookie
	oauthState := repo.generateStateOauthCookie(w)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the state query parameter on your redirect callback.
	*/
	u := googleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

func (repo *DBRepo) OauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Read oauthState from Cookie
	oauthState, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	data, err := repo.getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	var response models.GoogleUserResult

	err = json.Unmarshal(data, &response)
	if err != nil {
		log.Println(err)
		return
	}
	if response.Verified_email == false {
		repo.App.Session.Put(r.Context(), "error", "User must verify google account")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	user, err := repo.DB.GetUserById(response.Id)
	if err != nil {
		repo.App.Session.Put(r.Context(), "error", "User doesn't exist")
		repo.App.Session.Put(r.Context(), "user", response)
		http.Redirect(w, r, "/user/sign-up", http.StatusSeeOther)
		return
	}
	app.Session.Put(r.Context(), "user_id", user.ID)
	app.Session.Put(r.Context(), "user", user)
	repo.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/admin/dashboard", http.StatusSeeOther)
}

func (repo *DBRepo) generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func (repo *DBRepo) getUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

func (repo *DBRepo) validateGoogleJwt(jwt string) (bool, string) {
	response, err := http.Get(oauthGoogleUrlAPI + jwt)
	if err != nil {
		log.Println(fmt.Errorf("failed getting user info: %s", err.Error()))
		return false, ""
	}
	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(fmt.Errorf("failed read response: %s", err.Error()))
		return false, ""
	}
	log.Println(string(content))
	var g models.GoogleUserResult

	err = json.Unmarshal(content, &g)
	if err != nil {
		log.Println(err)
		return false, ""
	}
	log.Printf("%+v", g)
	_, err = repo.DB.GetUserById(g.Id)
	if err != nil {
		return false, ""
	}

	return true, g.Id
}
